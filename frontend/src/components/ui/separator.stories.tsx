import type { Meta, StoryObj } from "@storybook/react-vite"
import { Separator } from "@/components/ui/separator"

const meta = {
  title: "UI/Separator",
  component: Separator,
  tags: ["autodocs"],
} satisfies Meta<typeof Separator>

export default meta

type Story = StoryObj<typeof meta>

export const Horizontal: Story = {
  render: () => (
    <div className="w-72 space-y-3 text-sm">
      <p>商品信息</p>
      <Separator />
      <p>拼团规则</p>
    </div>
  ),
}

export const Vertical: Story = {
  render: () => (
    <div className="flex h-10 items-center gap-4 text-sm">
      <span>拼团价</span>
      <Separator orientation="vertical" />
      <span>¥199</span>
    </div>
  ),
}
