import type { Meta, StoryObj } from "@storybook/react-vite"
import {
  MerchantSignUpForm,
  SignUpForm,
} from "@/components/auth/sign-up-form"

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
  render: () => <MerchantSignUpForm />,
}
