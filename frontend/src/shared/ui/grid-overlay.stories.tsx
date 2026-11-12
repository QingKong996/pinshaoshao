import type { Meta, StoryObj } from "@storybook/react-vite"
import { GridOverlay } from "./grid-overlay"

const meta = {
  title: "Components/GridOverlay",
  component: GridOverlay,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof GridOverlay>

export default meta

type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <div className="min-h-svh bg-background">
      <GridOverlay />
    </div>
  ),
}
