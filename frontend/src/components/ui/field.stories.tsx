import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"

const meta = {
  title: "UI/Field",
  component: Field,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="w-72">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof Field>

export default meta

type Story = StoryObj<typeof meta>

export const WithDescription: Story = {
  render: (args) => (
    <Field {...args}>
      <FieldLabel htmlFor="field-username">用户名</FieldLabel>
      <Input id="field-username" placeholder="请输入用户名" />
      <FieldDescription>5-20 个字符，注册后不可修改。</FieldDescription>
    </Field>
  ),
}

export const WithError: Story = {
  render: () => (
    <Field data-invalid>
      <FieldLabel htmlFor="field-email">邮箱</FieldLabel>
      <Input id="field-email" aria-invalid defaultValue="not-an-email" />
      <FieldError>请输入有效的邮箱地址。</FieldError>
    </Field>
  ),
}

export const Group: Story = {
  render: () => (
    <FieldGroup>
      <Field>
        <FieldLabel htmlFor="field-signin-username">用户名</FieldLabel>
        <Input id="field-signin-username" placeholder="请输入用户名" />
      </Field>
      <Field>
        <FieldLabel htmlFor="field-signin-password">密码</FieldLabel>
        <Input id="field-signin-password" type="password" placeholder="请输入密码" />
      </Field>
    </FieldGroup>
  ),
}
