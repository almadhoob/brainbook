import type { AvatarProps } from '@nuxt/ui'

export interface User {
  id: number
  name: string
  email: string
  avatar?: AvatarProps
  location: string
}

export interface Message {
  id: number
  unread?: boolean
  from: User
  subject: string
  body: string
  date: string
}

export interface Notification {
  id: number
  unread?: boolean
  sender: User
  body: string
  date: string
}
