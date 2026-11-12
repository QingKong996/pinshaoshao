import type { Meta, StoryObj } from "@storybook/react-vite"
import { UserMenu } from "./user-menu"

const meta = {
  title: "Components/UserMenu",
  component: UserMenu,
  tags: ["autodocs"],
  args: {
    session: {
      token: "storybook-token",
      username: "alice01",
    },
  },
} satisfies Meta<typeof UserMenu>

export default meta

type Story = StoryObj<typeof meta>

export const Default: Story = {}
