import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Avatar,
  AvatarBadge,
  AvatarFallback,
  AvatarGroup,
  AvatarGroupCount,
} from "./avatar"

const meta = {
  title: "UI/Avatar",
  component: Avatar,
  tags: ["autodocs"],
} satisfies Meta<typeof Avatar>

export default meta

type Story = StoryObj<typeof meta>

export const Fallback: Story = {
  render: (args) => (
    <Avatar {...args}>
      <AvatarFallback>林</AvatarFallback>
    </Avatar>
  ),
}

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Avatar size="sm">
        <AvatarFallback>林</AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback>周</AvatarFallback>
      </Avatar>
      <Avatar size="lg">
        <AvatarFallback>陈</AvatarFallback>
      </Avatar>
    </div>
  ),
}

export const Group: Story = {
  render: () => (
    <AvatarGroup>
      <Avatar>
        <AvatarFallback className="bg-gradient-to-br from-orange-400 to-rose-500 font-medium text-white">
          林
        </AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback className="bg-gradient-to-br from-rose-400 to-pink-500 font-medium text-white">
          周
        </AvatarFallback>
      </Avatar>
      <Avatar>
        <AvatarFallback className="bg-gradient-to-br from-amber-400 to-orange-500 font-medium text-white">
          陈
        </AvatarFallback>
      </Avatar>
      <AvatarGroupCount>+3</AvatarGroupCount>
    </AvatarGroup>
  ),
}

export const WithBadge: Story = {
  render: () => (
    <Avatar size="lg">
      <AvatarFallback>林</AvatarFallback>
      <AvatarBadge />
    </Avatar>
  ),
}
