import type { Meta, StoryObj } from "@storybook/react-vite"
import { SignInForm } from "@/components/auth/sign-in-form"

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
