import type { Meta, StoryObj } from "@storybook/react-vite"
import { SiteFooter } from "@/components/site-footer"

const meta = {
  title: "Components/SiteFooter",
  component: SiteFooter,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof SiteFooter>

export default meta

type Story = StoryObj<typeof meta>

export const SignedOut: Story = {}
