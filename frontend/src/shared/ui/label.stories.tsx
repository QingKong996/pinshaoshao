import type { Meta, StoryObj } from "@storybook/react-vite"
import { Input } from "./input"
import { Label } from "./label"

const meta = {
  title: "UI/Label",
  component: Label,
  tags: ["autodocs"],
  args: {
    children: "用户名",
  },
} satisfies Meta<typeof Label>

export default meta

type Story = StoryObj<typeof meta>

export const Playground: Story = {}

export const WithInput: Story = {
  render: (args) => (
    <div className="grid w-72 gap-2">
      <Label {...args} htmlFor="label-with-input" />
      <Input id="label-with-input" placeholder="请输入用户名" />
    </div>
  ),
}
