export type LoginRequest = {
  username: string
  password: string
}

export type RegisterRequest = {
  username: string
  password: string
  email?: string
}

export type MerchantRegisterRequest = {
  username: string
  name: string
  password: string
  email: string
}

export type TokenResponse = {
  token: string
}

export type UserSession = {
  token: string
  username: string
}

export type MerchantSession = {
  token: string
  username: string
}

export type UserAuthContextValue = {
  session: UserSession | null
  signIn: (credentials: LoginRequest) => Promise<void>
  signUp: (details: RegisterRequest) => Promise<void>
  signOut: () => Promise<void>
}

export type MerchantAuthContextValue = {
  session: MerchantSession | null
  signIn: (credentials: LoginRequest) => Promise<void>
  signUp: (details: MerchantRegisterRequest) => Promise<void>
  signOut: () => Promise<void>
}
