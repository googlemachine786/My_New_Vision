"use client";

import Image from "next/image";
import { useState, useEffect } from "react";
import { useAuthStore } from "@/store/authStore";
import { useRouter } from "next/navigation";
import {
  LayoutDashboard,
  BookOpen,
  MessageSquare,
  Scissors,
  Search,
  Bell,
  LogOut,
  X,
} from "lucide-react";

const navItems = [
  { label: "Dashboard", icon: LayoutDashboard },
  { label: "Learn", icon: BookOpen },
  { label: "Ask", icon: MessageSquare },
  { label: "Practice", icon: Scissors },
];

export default function OnboardingShell({ children }: { children: React.ReactNode }) {
  const clear = useAuthStore((s) => s.clear);
  const router = useRouter();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarVisible, setSidebarVisible] = useState(false);

  function openSidebar() {
    setSidebarOpen(true);
    requestAnimationFrame(() => requestAnimationFrame(() => setSidebarVisible(true)));
  }

  function closeSidebar() {
    setSidebarVisible(false);
  }

  useEffect(() => {
    if (!sidebarVisible && sidebarOpen) {
      const t = setTimeout(() => setSidebarOpen(false), 300);
      return () => clearTimeout(t);
    }
  }, [sidebarVisible, sidebarOpen]);

  async function handleLogout() {
    await clear();
    router.replace("/login");
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 h-16 flex items-center px-10 gap-10 sticky top-0 z-10">
        {/* Hamburger */}
        <button
          className="text-gray-600 hover:text-gray-900 transition-colors shrink-0"
          onClick={() => openSidebar()}
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="3" y1="12" x2="21" y2="12" />
            <line x1="3" y1="18" x2="21" y2="18" />
          </svg>
        </button>

        {/* Logo */}
        <div>
          <Image src="/header-logo.png" alt="Visionary" width={120} height={32} className="shrink-0 h-7 w-auto" />
        </div>

        {/* Search */}
        <div className="flex-1 hidden sm:flex items-center relative">
          <input
            className="flex-1 text-sm outline-none text-gray-700 placeholder:text-gray-400 bg-white rounded-full px-5 py-2.5 pr-12 border border-gray-200"
            placeholder="What do you want to learn today?"
          />
          <div className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-800 pointer-events-none">
            <Search size={18} strokeWidth={2.5} />
          </div>
        </div>

        {/* Right actions */}
        <div className="ml-auto sm:ml-0 flex items-center gap-4">
          <button className="sm:hidden text-gray-500 hover:text-gray-700">
            <Search size={18} />
          </button>
          <button className="hidden sm:flex items-center gap-1.5 border border-gray-200 rounded-full px-3 py-1.5 text-xs text-gray-600 hover:bg-gray-50 transition-colors">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" /><line x1="2" y1="12" x2="22" y2="12" />
              <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
            </svg>
            ENG
          </button>
          <button className="text-gray-500 hover:text-gray-700">
            <Bell size={18} />
          </button>
          <button onClick={handleLogout} className="text-gray-500 hover:text-red-600 transition-colors" title="Logout">
            <LogOut size={18} />
          </button>
          <div className="w-8 h-8 rounded-full bg-gray-200 shrink-0" />
        </div>
      </header>

      <div className="flex flex-1 min-h-0">
        {/* Mobile sidebar drawer */}
        {sidebarOpen && (
          <div className="fixed inset-0 z-40 flex sm:hidden">
            {/* Backdrop */}
            <div
              className={`fixed inset-0 bg-black/40 transition-opacity duration-300 ${sidebarVisible ? "opacity-100" : "opacity-0"}`}
              onClick={closeSidebar}
            />
            {/* Drawer */}
            <aside
              className={`relative z-50 w-56 bg-white h-full flex flex-col py-6 px-4 gap-2 shadow-xl transition-transform duration-300 ease-in-out ${sidebarVisible ? "translate-x-0" : "-translate-x-full"}`}
            >
              <div className="flex items-center justify-between mb-4">
                <Image src="/header-logo.png" alt="Visionary" width={100} height={28} className="h-6 w-auto" />
                <button onClick={closeSidebar} className="text-gray-500 hover:text-gray-800 transition-colors">
                  <X size={20} />
                </button>
              </div>
              {navItems.map(({ label, icon: Icon }, i) => (
                <div
                  key={label}
                  title="Available after onboarding"
                  style={{ transitionDelay: sidebarVisible ? `${60 + i * 50}ms` : "0ms" }}
                  className={`flex items-center gap-3 px-3 py-2.5 rounded-xl opacity-40 cursor-not-allowed transition-all duration-300 ${sidebarVisible ? "opacity-40 translate-x-0" : "opacity-0 -translate-x-4"}`}
                >
                  <Icon size={18} className="text-gray-600 shrink-0" />
                  <span className="text-sm text-gray-500">{label}</span>
                </div>
              ))}
            </aside>
          </div>
        )}

        {/* Sidebar — hidden on mobile, visible sm+ */}
        <aside className="hidden sm:flex w-16 md:w-20 bg-white border-r border-gray-200 flex-col items-center py-6 gap-5 shrink-0">
          {navItems.map(({ label, icon: Icon }) => (
            <div
              key={label}
              className="flex flex-col items-center gap-1 opacity-40 cursor-not-allowed"
              title="Available after onboarding"
            >
              <div className="w-9 h-9 md:w-10 md:h-10 flex items-center justify-center rounded-xl">
                <Icon size={18} className="text-gray-600" />
              </div>
              <span className="text-[9px] md:text-[10px] text-gray-500">{label}</span>
            </div>
          ))}
        </aside>

        {/* Main */}
        <main className="flex-1 flex items-start sm:items-center justify-center p-4 sm:p-6 overflow-y-auto">
          {children}
        </main>
      </div>
    </div>
  );
}
