package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"pinshaoshao/internal/mock/logistics/model"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

const (
	maxMerchantNoLen  = 64
	maxItemSummaryLen = 200
	maxCarrierLen     = 64
	maxCityLen        = 64
	maxPartyNameLen   = 100
	maxPhoneLen       = 32
	maxAddressLen     = 500
	maxURLLen         = 2048
)

type carrierOption struct {
	Name string
	Code string
}

var carrierOptions = []carrierOption{
	{Name: "顺丰速运", Code: "SF"},
	{Name: "京东物流", Code: "JD"},
	{Name: "中通快递", Code: "ZTO"},
	{Name: "圆通速递", Code: "YTO"},
	{Name: "申通快递", Code: "STO"},
	{Name: "韵达快递", Code: "YD"},
	{Name: "极兔速递", Code: "JT"},
}

var cityOptions = []string{
	"上海",
	"北京",
	"广州",
	"深圳",
	"杭州",
	"南京",
	"成都",
	"武汉",
	"苏州",
	"宁波",
}

type ShipmentRepository interface {
	Create(ctx context.Context, shipment *model.Shipment) (*model.Shipment, bool, error)
	Get(ctx context.Context, shipmentNo string) (*model.Shipment, error)
	Transition(
		ctx context.Context,
		shipmentNo string,
		action model.Action,
		now time.Time,
		details model.EventDetails,
	) (*model.Shipment, error)
}

type CreateShipmentInput struct {
	MerchantOrderNo string
	ItemSummary     string
	Carrier         string
	SenderName      string
	SenderPhone     string
	SenderAddress   string
	ReceiverName    string
	ReceiverPhone   string
	ReceiverAddress string
	OriginCity      string
	DestinationCity string
	NotifyURL       string
}

type ShipmentService struct {
	repository ShipmentRepository
	now        func() time.Time
}

func NewShipmentService(repository ShipmentRepository) *ShipmentService {
	return &ShipmentService{
		repository: repository,
		now:        time.Now,
	}
}

func (s *ShipmentService) Create(
	ctx context.Context,
	input CreateShipmentInput,
) (*model.Shipment, bool, error) {
	normalized, requestHash, err := normalizeCreateInput(input)
	if err != nil {
		return nil, false, err
	}

	now := s.now().UTC()
	originCity := normalized.OriginCity
	if originCity == "" {
		originCity = randomCity("")
	}
	destinationCity := normalized.DestinationCity
	if destinationCity == "" {
		destinationCity = randomCity(originCity)
	}
	carrier := normalized.Carrier
	if carrier == "" {
		carrier = carrierOptions[gofakeit.IntRange(0, len(carrierOptions)-1)].Name
	}
	itemSummary := normalized.ItemSummary
	if itemSummary == "" {
		itemSummary = truncateRunes(gofakeit.ProductName(), maxItemSummaryLen)
	}

	shipment := &model.Shipment{
		ShipmentNo:      uuid.NewString(),
		MerchantOrderNo: normalized.MerchantOrderNo,
		RequestHash:     requestHash,
		TrackingNo:      newTrackingNo(carrier),
		Carrier:         carrier,
		Status:          model.StatusPendingOutbound,
		ItemSummary:     itemSummary,
		Sender:          newParty(normalized.SenderName, normalized.SenderPhone, normalized.SenderAddress, originCity),
		Receiver:        newParty(normalized.ReceiverName, normalized.ReceiverPhone, normalized.ReceiverAddress, destinationCity),
		OriginCity:      originCity,
		DestinationCity: destinationCity,
		NotifyURL:       normalized.NotifyURL,
		CurrentLocation: originCity + "商家仓",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	created, isNew, err := s.repository.Create(ctx, shipment)
	if err != nil {
		return nil, false, fmt.Errorf("service.CreateShipment: %w", err)
	}
	return created, isNew, nil
}

func (s *ShipmentService) Get(ctx context.Context, shipmentNo string) (*model.Shipment, error) {
	if _, err := parseShipmentNo(shipmentNo); err != nil {
		return nil, err
	}
	shipment, err := s.repository.Get(ctx, shipmentNo)
	if err != nil {
		return nil, fmt.Errorf("service.GetShipment: %w", err)
	}
	return shipment, nil
}

func (s *ShipmentService) Outbound(ctx context.Context, shipmentNo string) (*model.Shipment, error) {
	return s.transition(ctx, shipmentNo, model.ActionOutbound)
}

func (s *ShipmentService) Advance(ctx context.Context, shipmentNo string) (*model.Shipment, error) {
	return s.transition(ctx, shipmentNo, model.ActionAdvance)
}

func (s *ShipmentService) Complete(ctx context.Context, shipmentNo string) (*model.Shipment, error) {
	return s.transition(ctx, shipmentNo, model.ActionComplete)
}

func (s *ShipmentService) transition(
	ctx context.Context,
	shipmentNo string,
	action model.Action,
) (*model.Shipment, error) {
	if _, err := parseShipmentNo(shipmentNo); err != nil {
		return nil, err
	}

	shipment, err := s.repository.Get(ctx, shipmentNo)
	if err != nil {
		return nil, fmt.Errorf("service.TransitionShipment: %w", err)
	}
	if action == model.ActionOutbound && shipment.IsDispatched() {
		return shipment, nil
	}
	if action == model.ActionComplete && shipment.Status.IsTerminal() {
		return shipment, nil
	}

	transition, err := shipment.Next(action)
	if err != nil {
		return nil, err
	}
	details := eventDetails(*shipment, transition.Status)
	updated, err := s.repository.Transition(
		ctx,
		shipmentNo,
		action,
		s.now().UTC(),
		details,
	)
	if err != nil {
		return nil, fmt.Errorf("service.TransitionShipment: %w", err)
	}
	return updated, nil
}

func normalizeCreateInput(input CreateShipmentInput) (CreateShipmentInput, string, error) {
	input.MerchantOrderNo = strings.TrimSpace(input.MerchantOrderNo)
	input.ItemSummary = strings.TrimSpace(input.ItemSummary)
	input.Carrier = strings.TrimSpace(input.Carrier)
	input.SenderName = strings.TrimSpace(input.SenderName)
	input.SenderPhone = strings.TrimSpace(input.SenderPhone)
	input.SenderAddress = strings.TrimSpace(input.SenderAddress)
	input.ReceiverName = strings.TrimSpace(input.ReceiverName)
	input.ReceiverPhone = strings.TrimSpace(input.ReceiverPhone)
	input.ReceiverAddress = strings.TrimSpace(input.ReceiverAddress)
	input.OriginCity = strings.TrimSpace(input.OriginCity)
	input.DestinationCity = strings.TrimSpace(input.DestinationCity)
	input.NotifyURL = strings.TrimSpace(input.NotifyURL)

	switch {
	case input.MerchantOrderNo == "" || utf8.RuneCountInString(input.MerchantOrderNo) > maxMerchantNoLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.ItemSummary) > maxItemSummaryLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.Carrier) > maxCarrierLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.SenderName) > maxPartyNameLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.SenderPhone) > maxPhoneLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.SenderAddress) > maxAddressLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.ReceiverName) > maxPartyNameLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.ReceiverPhone) > maxPhoneLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.ReceiverAddress) > maxAddressLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.OriginCity) > maxCityLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	case utf8.RuneCountInString(input.DestinationCity) > maxCityLen:
		return CreateShipmentInput{}, "", model.ErrInvalidShipmentInput
	}
	if err := validateCallbackURL(input.NotifyURL); err != nil {
		return CreateShipmentInput{}, "", err
	}

	encoded, err := json.Marshal(input)
	if err != nil {
		return CreateShipmentInput{}, "", err
	}
	sum := sha256.Sum256(encoded)
	return input, hex.EncodeToString(sum[:]), nil
}

func validateCallbackURL(value string) error {
	if value == "" || len(value) > maxURLLen {
		return model.ErrInvalidShipmentInput
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return model.ErrInvalidShipmentInput
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return model.ErrInvalidShipmentInput
	}
	return nil
}

func parseShipmentNo(value string) (string, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil {
		return "", model.ErrInvalidShipmentNumber
	}
	return parsed.String(), nil
}

func newParty(name, phone, address, city string) model.Party {
	if name == "" {
		name = gofakeit.Name()
	}
	if phone == "" {
		phone = gofakeit.Phone()
	}
	if address == "" {
		address = fmt.Sprintf("%s%s, %s", gofakeit.StreetNumber(), gofakeit.Street(), city)
	}
	return model.Party{
		Name:    name,
		Phone:   phone,
		Address: address,
	}
}

func randomCity(exclude string) string {
	for {
		city := gofakeit.RandomString(cityOptions)
		if city != exclude {
			return city
		}
	}
}

func newTrackingNo(carrier string) string {
	code := "PKG"
	for _, option := range carrierOptions {
		if option.Name == carrier {
			code = option.Code
			break
		}
	}
	return code + gofakeit.Numerify("############")
}

func eventDetails(shipment model.Shipment, next model.Status) model.EventDetails {
	switch next {
	case model.StatusPendingPickup:
		return model.EventDetails{
			Description: "商家已出库，等待快递揽收",
			Location:    shipment.OriginCity + "商家仓",
		}
	case model.StatusPickedUp:
		return model.EventDetails{
			Description: fmt.Sprintf("快递员 %s 已揽收", gofakeit.Name()),
			Location:    shipment.OriginCity + "揽收点",
		}
	case model.StatusInTransit:
		return model.EventDetails{
			Description: fmt.Sprintf("快件已从%s发往%s", shipment.OriginCity, shipment.DestinationCity),
			Location:    shipment.OriginCity + "分拨中心",
		}
	case model.StatusAtTransitHub:
		return model.EventDetails{
			Description: fmt.Sprintf("快件已到达%s转运中心", shipment.DestinationCity),
			Location:    shipment.DestinationCity + "转运中心",
		}
	case model.StatusOutForDelivery:
		return model.EventDetails{
			Description: fmt.Sprintf("快递员 %s 正在派送", gofakeit.Name()),
			Location:    shipment.DestinationCity + "派送站",
		}
	case model.StatusDelivered:
		return model.EventDetails{
			Description: "快件已签收，感谢使用",
			Location:    shipment.DestinationCity,
		}
	default:
		return model.EventDetails{}
	}
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
