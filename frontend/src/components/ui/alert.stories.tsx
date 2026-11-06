import type { Meta, StoryObj } from "@storybook/react-vite"
import { CircleAlert, CircleCheck, Info } from "lucide-react"
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert"
import { Button } from "@/components/ui/button"

const meta = {
  title: "UI/Alert",
  component: Alert,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="w-96">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof Alert>

export default meta

type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: (args) => (
    <Alert {...args}>
      <Info />
      <AlertTitle>拼团进行中</AlertTitle>
      <AlertDescription>还差 1 人成团，24 小时内有效。</AlertDescription>
    </Alert>
  ),
}

export const Destructive: Story = {
  render: (args) => (
    <Alert {...args} variant="destructive">
      <CircleAlert />
      <AlertDescription>用户名或密码不正确，请重试。</AlertDescription>
    </Alert>
  ),
}

export const WithAction: Story = {
  render: () => (
    <Alert>
      <CircleCheck />
      <AlertTitle>订单已提交</AlertTitle>
      <AlertDescription>成团成功后 24 小时内发货。</AlertDescription>
      <AlertAction>
        <Button variant="outline" size="xs">
          查看订单
        </Button>
      </AlertAction>
    </Alert>
  ),
}
