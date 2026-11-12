import type { Meta, StoryObj } from "@storybook/react-vite"
import { MerchantSignInForm } from "./merchant-sign-in-form"
import { SignInForm } from "./sign-in-form"

const meta = {
  title: "Components/Auth/SignInForm",
  component: SignInForm,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div className="w-80">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof SignInForm>

export default meta

type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Merchant: Story = {
  render: () => <MerchantSignInForm redirectTo="/merchant/dashboard" />,
}
