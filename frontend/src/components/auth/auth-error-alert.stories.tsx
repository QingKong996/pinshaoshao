import type { Meta, StoryObj } from "@storybook/react-vite"
import { AuthErrorAlert } from "@/components/auth/auth-error-alert"

const meta = {
  title: "Components/Auth/AuthErrorAlert",
  component: AuthErrorAlert,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="w-96">
        <Story />
      </div>
    ),
  ],
  args: {
    message: "用户名或密码不正确，请重试。",
  },
} satisfies Meta<typeof AuthErrorAlert>

export default meta

type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const LongMessage: Story = {
  args: {
    message:
      "注册失败：该用户名已被占用，请更换用户名后重试。如果问题持续存在，请稍后再试或联系客服。",
  },
}
