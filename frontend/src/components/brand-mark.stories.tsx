import type { Meta, StoryObj } from "@storybook/react-vite"
import { BrandMark } from "@/components/brand-mark"

const meta = {
  title: "Components/BrandMark",
  component: BrandMark,
  tags: ["autodocs"],
} satisfies Meta<typeof BrandMark>

export default meta

type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: (args) => <BrandMark {...args} />,
}

export const Inverted: Story = {
  render: () => (
    <div className="rounded-2xl bg-neutral-950 p-8">
      <BrandMark inverted />
    </div>
  ),
}
