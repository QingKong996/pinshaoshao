import type { Meta, StoryObj } from "@storybook/react-vite"
import { ArrowRight, Plus } from "lucide-react"
import { Button } from "@/components/ui/button"

const VARIANT_LABELS = {
  default: "默认",
  outline: "描边",
  secondary: "次要",
  ghost: "幽灵",
  destructive: "危险",
  link: "链接",
} as const

const SIZE_LABELS = {
  xs: "超小",
  sm: "小",
  default: "默认",
  lg: "大",
} as const

const meta = {
  title: "UI/Button",
  component: Button,
  tags: ["autodocs"],
  args: {
    children: "立即拼团",
    variant: "default",
    size: "default",
  },
  argTypes: {
    variant: {
      control: "select",
      options: Object.keys(VARIANT_LABELS),
    },
    size: {
      control: "select",
      options: ["default", "xs", "sm", "lg", "icon", "icon-xs", "icon-sm", "icon-lg"],
    },
  },
} satisfies Meta<typeof Button>

export default meta

type Story = StoryObj<typeof meta>

export const Playground: Story = {}

export const Variants: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      {(Object.keys(VARIANT_LABELS) as Array<keyof typeof VARIANT_LABELS>).map(
        (variant) => (
          <Button key={variant} variant={variant}>
            {VARIANT_LABELS[variant]}
          </Button>
        ),
      )}
    </div>
  ),
}

export const Sizes: Story = {
  render: () => (
    <div className="flex flex-wrap items-end gap-3">
      {(Object.keys(SIZE_LABELS) as Array<keyof typeof SIZE_LABELS>).map((size) => (
        <Button key={size} size={size}>
          {SIZE_LABELS[size]}
        </Button>
      ))}
    </div>
  ),
}

export const WithIcons: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      <Button>
        <Plus data-icon="inline-start" />
        发起拼团
      </Button>
      <Button variant="outline">
        去结算
        <ArrowRight data-icon="inline-end" />
      </Button>
      <Button variant="outline" size="icon" aria-label="发起拼团">
        <Plus />
      </Button>
    </div>
  ),
}

export const Disabled: Story = {
  args: {
    disabled: true,
  },
}
