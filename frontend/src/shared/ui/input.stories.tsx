import type { Meta, StoryObj } from "@storybook/react-vite"
import { Input } from "./input"

const meta = {
  title: "UI/Input",
  component: Input,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="w-72">
        <Story />
      </div>
    ),
  ],
  args: {
    placeholder: "请输入用户名",
  },
} satisfies Meta<typeof Input>

export default meta

type Story = StoryObj<typeof meta>

export const Playground: Story = {}

export const WithValue: Story = {
  args: {
    defaultValue: "alice01",
  },
}

export const Invalid: Story = {
  args: {
    "aria-invalid": true,
    defaultValue: "a",
  },
}

export const Disabled: Story = {
  args: {
    disabled: true,
    defaultValue: "不可编辑",
  },
}
