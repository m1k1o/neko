export interface Settings {
  can_send: boolean
  can_receive: boolean
}

export const CHAT_INIT = "chat/init"
export const CHAT_MESSAGE = "chat/message"

export interface Init {
	enabled: boolean
}

export interface Content {
  text: string
}

export interface Message {
  id: string
  created: Date
  content: Content
}
