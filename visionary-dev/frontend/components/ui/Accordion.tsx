'use client'
import { useState } from 'react'
import { Minus, Plus } from 'lucide-react'

export interface AccordionItem {
  id: string
  title: string
  content: React.ReactNode
  onClick?: () => void
}

interface Props {
  items: AccordionItem[]
  defaultOpen?: string[]
  allowMultiple?: boolean
  activeId?: string
}

export default function Accordion({ items, defaultOpen = [], allowMultiple = false, activeId }: Props) {
  const [openIds, setOpenIds] = useState<string[]>(defaultOpen)

  function toggle(id: string) {
    setOpenIds((prev) => {
      const isOpen = prev.includes(id)
      if (allowMultiple) {
        return isOpen ? prev.filter((i) => i !== id) : [...prev, id]
      }
      return isOpen ? [] : [id]
    })
  }

  return (
    <div className="rounded-xl border border-gray-200 overflow-hidden shadow-sm">
      {items.map((item, index) => {
        const isOpen = openIds.includes(item.id)
        const isActive = activeId === item.id
        const isLast = index === items.length - 1

        return (
          <div key={item.id}>
            <h2>
              <button
                type="button"
                onClick={() => {
                  toggle(item.id)
                  item.onClick?.()
                }}
                aria-expanded={isOpen}
                aria-controls={`accordion-body-${item.id}`}
                className={`flex items-center justify-between w-full px-5 py-4 font-medium text-left gap-3 hover:text-gray-900 hover:bg-gray-50 transition-colors ${
                  !isLast || isOpen ? 'border-b border-gray-200' : ''
                } ${isActive ? 'text-gray-900 font-semibold' : 'text-gray-700'}`}
                style={{ fontSize: '16px' }}
              >
                <span>{item.title}</span>
                {isOpen
                  ? <Minus size={14} className="shrink-0 text-gray-400" />
                  : <Plus size={14} className="shrink-0 text-gray-400" />
                }
              </button>
            </h2>

            <div
              id={`accordion-body-${item.id}`}
              role="region"
              aria-labelledby={`accordion-heading-${item.id}`}
              className={`overflow-hidden transition-all duration-200 ${isOpen ? 'max-h-[1000px] opacity-100' : 'max-h-0 opacity-0'} ${!isLast ? 'border-b border-gray-200' : ''}`}
            >
              <div className="px-5 py-4 leading-relaxed" style={{ fontSize: '14px', color: '#6b7280' }}>
                {item.content}
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
