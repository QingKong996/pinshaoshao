import type { Meta, StoryObj } from "@storybook/react-vite"
import { PasswordInput } from "./password-input"

const meta = {
  title: "Components/Auth/PasswordInput",
  component: PasswordInput,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="w-72">
        <Story />
      </div>
    ),
  ],
  args: {
    placeholder: "请输入密码",
  },
} satisfies Meta<typeof PasswordInput>

export default meta

type Story = StoryObj<typeof meta>

export const Playground: Story = {}

export const WithValue: Story = {
  args: {
    defaultValue: "pinshaoshao123",
  },
}

export const Disabled: Story = {
  args: {
    disabled: true,
    defaultValue: "不可编辑",
  },
}

export const Invalid: Story = {
  args: {
    "aria-invalid": true,
    defaultValue: "123",
  },
}
