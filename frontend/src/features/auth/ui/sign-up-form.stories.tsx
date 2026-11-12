import type { Meta, StoryObj } from "@storybook/react-vite"
import { MerchantSignUpForm } from "./merchant-sign-up-form"
import { SignUpForm } from "./sign-up-form"

const meta = {
  title: "Components/Auth/SignUpForm",
  component: SignUpForm,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="w-80">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof SignUpForm>

export default meta

type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Merchant: Story = {
  render: () => <MerchantSignUpForm redirectTo="/merchant/dashboard" />,
}
