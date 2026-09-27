'use client'
import { useEffect } from 'react'
import { Bookmark, ChevronDown, Minimize2, Paperclip } from 'lucide-react'
import Image from 'next/image'

interface Props {
  open: boolean
  onClose: () => void
  title?: string
  children: React.ReactNode
}

export default function Dialog({ open, onClose, title, children }: Props) {
  useEffect(() => {
    document.body.style.overflow = open ? 'hidden' : 'unset'
    return () => { document.body.style.overflow = 'unset' }
  }, [open])

  useEffect(() => {
    function handleEscape(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    if (open) window.addEventListener('keydown', handleEscape)
    return () => window.removeEventListener('keydown', handleEscape)
  }, [open, onClose])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="bg-white flex flex-col rounded-2xl shadow-2xl" style={{ width: '92vw', height: '92vh' }}>

      {/* Top bar */}
      <div className="flex items-center justify-between px-6 py-2 border-b border-gray-100 shrink-0">
        {/* Left — section title */}
        <span className="text-sm text-gray-500">{title}</span>

        {/* Center — ESC pill */}
        <div className="absolute left-1/2 -translate-x-1/2">
          <div className="flex items-center gap-2 bg-gray-900 text-white text-xs px-4 py-2 rounded-full">
            Press
            <kbd className="px-1.5 py-0.5 rounded border border-gray-600 bg-gray-700 font-mono text-xs">esc</kbd>
            to exit full screen
          </div>
        </div>

        {/* Right — actions */}
        <div className="flex items-center gap-3">
          <button className="text-gray-400 hover:text-gray-600 transition-colors">
            <Bookmark size={18} />
          </button>
          <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-full border border-gray-200 text-sm text-gray-600 hover:bg-gray-50 transition-colors">
            Text
            <ChevronDown size={14} />
          </button>
        </div>
      </div>

      {/* Scrollable content */}
      <div className="flex-1 overflow-y-auto px-8 py-6 text-sm text-gray-700 leading-relaxed">
        {children}
      </div>

      {/* Bottom input bar */}
      <div className="shrink-0 px-6 pb-2 pt-2 bg-white rounded-b-2xl">
        <div className="flex items-center justify-between bg-white border border-gray-200 rounded-full px-4 py-3 shadow-[0_-2px_16px_rgba(0,0,0,0.06)]">
          <div className="flex items-center gap-2 text-gray-400 flex-1">
            <Paperclip size={16} />
            <span className="text-sm">I have a doubt</span>
          </div>
          <button className="w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center hover:bg-blue-700 transition-colors shrink-0">
            <Image src="/voice.png" alt="voice" width={18} height={18} />
          </button>
        </div>
        {/* Minimize icon — bottom right outside input */}
        <div className="flex justify-end mt-1.5">
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <Minimize2 size={14} />
          </button>
        </div>
      </div>

    </div>
    </div>
  )
}
