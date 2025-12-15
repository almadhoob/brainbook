export interface ApiUserListItem {
  user_id: number
  user_full_name: string
  user_avatar?: string | null
  last_message_time?: string | null
}

export interface ApiConversationMessage {
  sender_id: number
  content: string
  created_at: string
}

export interface ReceiveMessageEventPayload {
  message: string
  sender_id: number
  receiver_id: number
  sent_at: string
}

export interface ReceiveGroupMessageEventPayload {
  message: string
  sender_id: number
  group_id: number
  sent_at: string
}

export interface WebsocketNotificationPayload {
  id: number
  type: string
  payload: Record<string, unknown> | null
  is_read: boolean
  created_at: string
}

export interface UiNotification {
  id: number
  type: string
  isRead: boolean
  createdAt: string
  payload: Record<string, unknown> | null
}
