'use client'
import SubjectDetail from '@/features/learn/detail'
import { use } from 'react'

export default function SubjectDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  return <SubjectDetail subjectId={id} />
}