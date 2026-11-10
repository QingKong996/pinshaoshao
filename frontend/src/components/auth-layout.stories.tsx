import type { Meta, StoryObj } from "@storybook/react-vite"
import { Link } from "react-router"
import { AuthLayout } from "@/components/auth-layout"
import { MerchantSignInForm } from "@/components/auth/merchant-sign-in-form"
import { MerchantSignUpForm } from "@/components/auth/merchant-sign-up-form"
import { SignInForm } from "@/components/auth/sign-in-form"
import { SignUpForm } from "@/components/auth/sign-up-form"

const meta = {
  title: "Components/AuthLayout",
  component: AuthLayout,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
  },
  args: {
    title: "登录",
    description: "欢迎回来，登录后继续拼团。",
    footer: (
      <>
        还没有账号？
        <Link
          to="/signup"
          className="font-medium text-foreground underline-offset-4 transition-colors hover:underline"
        >
          免费注册
        </Link>
      </>
    ),
    children: <SignInForm />,
  },
} satisfies Meta<typeof AuthLayout>

export default meta

type Story = StoryObj<typeof meta>

export const SignIn: Story = {}

export const SignUp: Story = {
  args: {
    title: "免费注册",
    description: "创建账号，开启拼团省钱之旅。",
    footer: (
      <>
        已有账号？
        <Link
          to="/signin"
          className="font-medium text-foreground underline-offset-4 transition-colors hover:underline"
        >
          直接登录
        </Link>
      </>
    ),
    children: <SignUpForm />,
  },
}

export const MerchantSignIn: Story = {
  args: {
    variant: "merchant",
    title: "商家登录",
    description: "登录店铺，管理你的商品与订单。",
    footer: (
      <>
        还没有店铺？
        <Link
          to="/merchant/signup"
          className="font-medium text-foreground underline-offset-4 transition-colors hover:underline"
        >
          立即入驻
        </Link>
      </>
    ),
    children: <MerchantSignInForm />,
  },
}

export const MerchantSignUp: Story = {
  args: {
    variant: "merchant",
    title: "开店入驻",
    description: "注册商家账号，把好货卖给拼团买家。",
    footer: (
      <>
        已有店铺？
        <Link
          to="/merchant/signin"
          className="font-medium text-foreground underline-offset-4 transition-colors hover:underline"
        >
          去登录
        </Link>
      </>
    ),
    children: <MerchantSignUpForm />,
  },
}
