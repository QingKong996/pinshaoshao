import type { Preview } from "@storybook/react-vite"
import { MemoryRouter } from "react-router"
import "@/i18n"
import { AuthProvider } from "@/lib/auth"
import "../src/index.css"

const preview: Preview = {
  decorators: [
    (Story) => (
      <MemoryRouter>
        <AuthProvider>
          <Story />
        </AuthProvider>
      </MemoryRouter>
    ),
  ],
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    layout: "centered",
  },
}

export default preview
