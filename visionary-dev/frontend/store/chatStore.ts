import { create } from "zustand";
import type { Chat, ChatMessage } from "@/types";

interface ChatState {
  chats: Chat[];
  activeChatId: string | null;
  streaming: boolean;
  setChats: (chats: Chat[]) => void;
  setActiveChat: (id: string) => void;
  addMessage: (chatId: string, message: ChatMessage) => void;
  appendToLastMessage: (chatId: string, chunk: string) => void;
  setStreaming: (value: boolean) => void;
}

export const useChatStore = create<ChatState>((set) => ({
  chats: [],
  activeChatId: null,
  streaming: false,

  setChats: (chats) => set({ chats }),

  setActiveChat: (id) => set({ activeChatId: id }),

  addMessage: (chatId, message) =>
    set((state) => ({
      chats: state.chats.map((c) =>
        c.id === chatId ? { ...c, messages: [...c.messages, message] } : c
      ),
    })),

  appendToLastMessage: (chatId, chunk) =>
    set((state) => ({
      chats: state.chats.map((c) => {
        if (c.id !== chatId) return c;
        const messages = [...c.messages];
        if (messages.length === 0) return c;
        const last = messages[messages.length - 1];
        messages[messages.length - 1] = {
          ...last,
          content: last.content + chunk,
        };
        return { ...c, messages };
      }),
    })),

  setStreaming: (value) => set({ streaming: value }),
}));
