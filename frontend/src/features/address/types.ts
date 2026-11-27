export type Address = {
  id: number
  receiver_name: string
  receiver_phone: string
  shipping_address: string
  is_default: boolean
  created_at: string
  updated_at: string
}

export type AddressInput = {
  receiver_name: string
  receiver_phone: string
  shipping_address: string
  is_default: boolean
}
