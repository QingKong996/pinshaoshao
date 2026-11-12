import type { Preview } from "@storybook/react-vite"
import { MemoryRouter } from "react-router"
import "@/app/i18n"
import { AppProviders } from "@/app/providers"
import "../src/app/styles.css"

const preview: Preview = {
  decorators: [
    (Story) => (
      <MemoryRouter>
        <AppProviders>
          <Story />
        </AppProviders>
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
