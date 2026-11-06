import type { Meta, StoryObj } from "@storybook/react-vite"
import { SiteHeader } from "@/components/site-header"
import { AuthContext, type AuthContextValue } from "@/lib/auth-context"

const MOCK_AUTH: AuthContextValue = {
  session: {
    token: "storybook-token",
    username: "alice01",
  },
  signIn: async () => {},
  signUp: async () => {},
  signOut: async () => {},
}

const meta = {
  title: "Components/SiteHeader",
  component: SiteHeader,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof SiteHeader>

export default meta

type Story = StoryObj<typeof meta>

export const SignedOut: Story = {}

export const SignedIn: Story = {
  decorators: [
    (Story) => (
      <AuthContext.Provider value={MOCK_AUTH}>
        <Story />
      </AuthContext.Provider>
    ),
  ],
}
