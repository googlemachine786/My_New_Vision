'use client'
import SubjectNavbar from '@/features/learn/detail/components/SubjectNavbar'
import Accordion, { AccordionItem } from '@/components/ui/Accordion'
import Dialog from '@/components/ui/Dialog'
import { Fragment, use, useEffect, useState } from 'react'
import Image from 'next/image'
import { Bookmark, ChevronDown, Paperclip, Maximize2 } from 'lucide-react'
import { getChapterTopics } from '@/lib/auth'
import type { Topic, Subtopic } from '@/types'

interface Props {
  params: Promise<{
    id: string
    slug: string[]
  }>
}

const ChaptersSections = ({ params }: Props) => {
  const { id, slug } = use(params)
  const chapterId = slug[0]

  const [topics, setTopics] = useState<Topic[]>([])
  const [activeTopic, setActiveTopic] = useState<Topic | null>(null)
  const [activeSubtopic, setActiveSubtopic] = useState<Subtopic | null>(null)
  const [tocCollapsed, setTocCollapsed] = useState(false)
  const [fullscreen, setFullscreen] = useState(false)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!chapterId) return
    setLoading(true)
    getChapterTopics(chapterId)
      .then((data) => {
        setTopics(data)
        if (data.length > 0) setActiveTopic(data[0])
      })
      .catch(() => setTopics([]))
      .finally(() => setLoading(false))
  }, [chapterId])

  function scrollToSubtopic(subtopicId: string) {
    setTimeout(() => {
      const container = document.getElementById('content-scroll')
      const element = document.getElementById(subtopicId)
      if (container && element) {
        const top = element.offsetTop - container.offsetTop
        container.scrollTo({ top, behavior: 'smooth' })
      }
    }, 100)
  }

  const accordionItems: AccordionItem[] = topics.map((topic) => ({
    id: topic.pointNumber,
    title: `${topic.pointNumber} ${topic.title}`,
    onClick: () => {
      setActiveTopic(topic)
      setActiveSubtopic(null)
    },
    content: topic.subtopics.length > 0 ? (
      <div className="pl-4 border-l-2 border-blue-500 -ml-1">
        {topic.subtopics.map((sub) => (
          <button
            key={sub.id}
            onClick={() => {
              setActiveTopic(topic)
              setActiveSubtopic(sub)
              scrollToSubtopic(sub.id)
            }}
            className="w-full text-left px-3 py-2 text-gray-500 hover:text-gray-700 transition-colors"
            style={{ fontSize: '14px' }}
          >
            {sub.pointNumber} {sub.title}
          </button>
        ))}
      </div>
    ) : null,
  }))

  const contentBody = (
    <>
      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-4 bg-gray-100 rounded animate-pulse" />
          ))}
        </div>
      ) : !activeTopic ? (
        <p className="text-gray-400 text-sm">No topics found for this chapter.</p>
      ) : (
        <>
          <h3 className="text-base font-semibold text-gray-900 mb-4">
            {activeTopic.pointNumber} {activeTopic.title}
          </h3>

          {activeTopic.subtopics.length > 0 && (
            <div className="space-y-6 border-t border-gray-100 pt-4">
              {activeTopic.subtopics.map((sub) => (
                <div
                  key={sub.id}
                  id={sub.id}
                  className={`pb-4 border-b border-gray-100 last:border-0 transition-colors ${
                    activeSubtopic?.id === sub.id ? 'border-blue-100' : ''
                  }`}
                >
                  <div className={`flex items-center gap-2 mb-2 pb-2 border-b ${
                    activeSubtopic?.id === sub.id ? 'border-blue-300' : 'border-gray-100'
                  }`}>
                    <span className={`text-xs font-bold px-2 py-0.5 rounded-full ${
                      activeSubtopic?.id === sub.id
                        ? 'bg-blue-100 text-blue-600'
                        : 'bg-gray-100 text-gray-500'
                    }`}>
                      {sub.pointNumber}
                    </span>
                    <h4 className={`text-sm font-semibold ${
                      activeSubtopic?.id === sub.id ? 'text-blue-700' : 'text-gray-900'
                    }`}>
                      {sub.title}
                    </h4>
                  </div>
                  <p className="text-sm text-gray-500 leading-relaxed">
                    Content for {sub.pointNumber} will appear here.
                  </p>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </>
  )

  return (
    <Fragment>
      <SubjectNavbar />

      {/* Chapter Header */}
      <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
        <h2 className="text-lg font-semibold text-gray-900">
          {activeTopic ? activeTopic.title : 'Loading...'}
        </h2>
        <button className="flex items-center gap-2 px-4 py-2 rounded-full border border-gray-200 bg-white hover:bg-gray-50 transition-colors">
          <Image src="/pdf-icon.png" alt="PDF" width={20} height={20} />
          <span className="text-sm text-gray-700 font-medium">Academic Chapter</span>
        </button>
      </div>

      {/* Main Layout */}
      <div className="flex gap-4 p-4 bg-gray-50 min-h-screen items-start">

        {/* Left — Content Panel */}
        <div className="flex-1 flex flex-col gap-3 min-w-0">
          <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">

            {/* Panel Header */}
            <div className="flex items-center justify-between px-5 py-4 border-b border-gray-100">
              <span className="text-sm font-light text-gray-700">
                {activeTopic
                  ? `${activeTopic.pointNumber} ${activeTopic.title}`
                  : 'Select a topic'}
              </span>
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

            {/* Content — scrollable */}
            <div
              className="px-5 py-4 text-sm text-gray-700 leading-relaxed max-h-[420px] overflow-y-auto"
              id="content-scroll"
            >
              {contentBody}
            </div>

            {/* Action Buttons */}
            <div className="flex items-center justify-between px-5 py-4 border-t border-gray-100">
              <div className="flex items-center gap-3">
                <button
                  className="border border-gray-300 text-sm font-medium text-blue-500 hover:bg-blue-50 transition-colors"
                  style={{ width: 220, height: 48, borderRadius: 24, gap: 8 }}
                >
                  Simplify
                </button>
                <button
                  className="bg-blue-600 text-white text-sm font-medium hover:bg-blue-700 transition-colors"
                  style={{ width: 220, height: 48, borderRadius: 24, gap: 8 }}
                >
                  Test My Understanding
                </button>
              </div>
              <button onClick={() => setFullscreen(true)} className="text-gray-400 hover:text-gray-600 transition-colors">
                <Maximize2 size={16} />
              </button>
            </div>
          </div>

          {/* Bottom Input */}
          <div className="flex items-center justify-between bg-white rounded-2xl border border-gray-100 px-4 py-3">
            <div className="flex items-center gap-2 text-gray-400">
              <Paperclip size={16} />
              <span className="text-sm">I have a doubt</span>
            </div>
            <button className="w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center hover:bg-blue-700 transition-colors">
              <Image src="/voice.png" alt="voice" width={18} height={18} />
            </button>
          </div>
        </div>

        {/* Right — Table of Contents */}
        <div
          className="shrink-0 bg-white rounded-2xl border border-gray-100 flex flex-col transition-all duration-300"
          style={{ maxHeight: 'calc(100vh - 180px)', width: tocCollapsed ? 56 : 288 }}
        >
          {/* TOC Header */}
          <div className="px-4 py-4 border-b border-gray-100 shrink-0 flex justify-start">
            <button onClick={() => setTocCollapsed((v) => !v)}>
              <Image src="/sidebar.png" alt="sidebar" width={20} height={20} />
            </button>
          </div>

          {/* TOC List */}
          <div
            className="flex-1 overflow-y-auto"
            style={{ scrollbarWidth: 'thin', scrollbarColor: '#e5e7eb transparent' }}
          >
            {loading ? (
              <div className="p-4 space-y-3">
                {[1, 2, 3, 4].map((i) => (
                  <div key={i} className="h-4 bg-gray-100 rounded animate-pulse" />
                ))}
              </div>
            ) : tocCollapsed ? (
              /* Collapsed — just point numbers */
              <div className="flex flex-col items-center py-2">
                {topics.map((topic) => (
                  <button
                    key={topic.id}
                    onClick={() => {
                      setActiveTopic(topic)
                      setActiveSubtopic(null)
                    }}
                    className="w-full py-2.5 text-center text-sm transition-colors hover:text-gray-900"
                    style={{
                      color: activeTopic?.id === topic.id ? '#111827' : '#9ca3af',
                      fontWeight: activeTopic?.id === topic.id ? 700 : 400,
                    }}
                  >
                    {topic.pointNumber}
                  </button>
                ))}
              </div>
            ) : (
              /* Expanded — full accordion */
              <Accordion
                items={accordionItems}
                defaultOpen={topics.length > 0 ? [topics[0].pointNumber] : []}
                allowMultiple
                activeId={activeTopic?.pointNumber}
              />
            )}
          </div>
        </div>

      </div>

      {/* Fullscreen Dialog */}
      <Dialog
        open={fullscreen}
        onClose={() => setFullscreen(false)}
        title={activeTopic ? `${activeTopic.pointNumber} ${activeTopic.title}` : ''}
      >
        <div className="text-sm text-gray-700 leading-relaxed">
          {contentBody}
        </div>
      </Dialog>

    </Fragment>
  )
}

export default ChaptersSections
