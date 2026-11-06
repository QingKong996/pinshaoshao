import { useState } from "react"
import type { Meta, StoryObj } from "@storybook/react-vite"
import { LogOut, Settings, User } from "lucide-react"
import { buttonVariants } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"

const TRIGGER_CLASS = cn(buttonVariants({ variant: "outline" }), "cursor-pointer")

const LANGUAGE_OPTIONS = [
  { value: "zh-CN", label: "中文" },
  { value: "en", label: "English" },
] as const

const meta = {
  title: "UI/DropdownMenu",
  component: DropdownMenu,
  tags: ["autodocs"],
} satisfies Meta<typeof DropdownMenu>

export default meta

type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <DropdownMenu>
      <DropdownMenuTrigger className={TRIGGER_CLASS}>打开菜单</DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-48">
        <DropdownMenuLabel>我的账号</DropdownMenuLabel>
        <DropdownMenuGroup>
          <DropdownMenuItem>
            <User />
            个人资料
          </DropdownMenuItem>
          <DropdownMenuItem>
            <Settings />
            账号设置
            <DropdownMenuShortcut>⌘,</DropdownMenuShortcut>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuSub>
          <DropdownMenuSubTrigger>更多操作</DropdownMenuSubTrigger>
          <DropdownMenuSubContent>
            <DropdownMenuItem>导出数据</DropdownMenuItem>
            <DropdownMenuItem>打印发票</DropdownMenuItem>
          </DropdownMenuSubContent>
        </DropdownMenuSub>
        <DropdownMenuSeparator />
        <DropdownMenuItem variant="destructive">
          <LogOut />
          退出登录
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  ),
}

export const WithRadioGroup: Story = {
  render: () => (
    <DropdownMenu>
      <DropdownMenuTrigger className={TRIGGER_CLASS}>切换语言</DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-36">
        <DropdownMenuRadioGroup defaultValue="zh-CN">
          {LANGUAGE_OPTIONS.map((language) => (
            <DropdownMenuRadioItem
              key={language.value}
              value={language.value}
              closeOnClick
            >
              {language.label}
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  ),
}

function CheckboxItemsExample() {
  const [poolReminder, setPoolReminder] = useState(true)
  const [promoNotice, setPromoNotice] = useState(false)

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className={TRIGGER_CLASS}>通知设置</DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-44">
        <DropdownMenuCheckboxItem
          checked={poolReminder}
          onCheckedChange={setPoolReminder}
          closeOnClick={false}
        >
          拼团提醒
        </DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem
          checked={promoNotice}
          onCheckedChange={setPromoNotice}
          closeOnClick={false}
        >
          优惠通知
        </DropdownMenuCheckboxItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export const WithCheckboxItems: Story = {
  render: () => <CheckboxItemsExample />,
}
