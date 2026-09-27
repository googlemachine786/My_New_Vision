"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuthStore } from "@/store/authStore";
import { useState, useEffect } from "react";
import {
  LayoutDashboard,
  BookOpen,
  MessageSquare,
  Scissors,
  Search,
  Bell,
  LogOut,
  User,
  X,
} from "lucide-react";
import { cn } from "@/lib/utils";
import Popover, { PopoverHeader, PopoverItem } from "@/components/ui/Popover";

const navItems = [
  { label: "Dashboard", icon: LayoutDashboard, href: "/dashboard" },
  { label: "Learn", icon: BookOpen, href: "/learn" },
  { label: "Ask", icon: MessageSquare, href: "/ask" },
  { label: "Practice", icon: Scissors, href: "/practice" },
];

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { clear, user } = useAuthStore();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarVisible, setSidebarVisible] = useState(false);
  const [desktopExpanded, setDesktopExpanded] = useState(false);

  function openSidebar() {
    setSidebarOpen(true);
    // tiny delay so the element is mounted before transition starts
    requestAnimationFrame(() => requestAnimationFrame(() => setSidebarVisible(true)));
  }

  function closeSidebar() {
    setSidebarVisible(false);
  }

  // unmount after exit transition (300ms)
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
    <div className="h-screen bg-gray-50 flex flex-col overflow-hidden">
      {/* Header — no hamburger here */}
      <header className="bg-white border-b border-gray-200 h-16 flex items-center px-7 gap-10 sticky top-0 z-10">
        {/* Hamburger */}
        <button
          className="text-gray-600 hover:text-gray-900 transition-colors shrink-0"
          onClick={() => {
            // on mobile open drawer, on desktop toggle expand
            if (window.innerWidth < 640) openSidebar();
            else setDesktopExpanded((v) => !v);
          }}
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
          {/* Avatar + popover */}
          <Popover
            align="right"
            side="bottom"
            className="w-52"
            trigger={
              <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-xs font-semibold text-white shrink-0 hover:ring-2 hover:ring-blue-300 transition-all">
                {user?.fullName?.[0]?.toUpperCase() ?? "U"}
              </div>
            }
          >
            <PopoverHeader
              title={
                <div>
                  <p className="text-sm font-semibold text-gray-900 truncate">
                    {user?.fullName ?? "User"}
                  </p>
                  <p className="text-xs text-gray-500 truncate">{user?.email ?? ""}</p>
                </div>
              }
            />
            <div className="py-1">
              <PopoverItem icon={<User size={15} />} onClick={() => router.push("/profile")}>
                Profile
              </PopoverItem>
              <PopoverItem
                icon={<LogOut size={15} />}
                variant="danger"
                onClick={handleLogout}
              >
                Logout
              </PopoverItem>
            </div>
          </Popover>
        </div>
      </header>

      <div className="flex flex-1 min-h-0">
        {/* Mobile sidebar drawer */}
        {sidebarOpen && (
          <div className="fixed inset-0 z-40 flex sm:hidden">
            {/* Backdrop */}
            <div
              className={cn(
                "fixed inset-0 bg-black/40 transition-opacity duration-300",
                sidebarVisible ? "opacity-100" : "opacity-0"
              )}
              onClick={closeSidebar}
            />
            {/* Drawer */}
            <aside
              className={cn(
                "relative z-50 w-56 bg-white h-full flex flex-col py-6 px-4 gap-2 shadow-xl",
                "transition-transform duration-300 ease-in-out",
                sidebarVisible ? "translate-x-0" : "-translate-x-full"
              )}
            >
              <div className="flex items-center justify-between mb-4">
                <Image src="/header-logo.png" alt="Visionary" width={100} height={28} className="h-6 w-auto" />
                <button onClick={closeSidebar} className="text-gray-500 hover:text-gray-800 transition-colors">
                  <X size={20} />
                </button>
              </div>
              {navItems.map(({ label, icon: Icon, href }, i) => {
                const active = pathname === href;
                return (
                  <div
                    key={label}
                    style={{ transitionDelay: sidebarVisible ? `${60 + i * 50}ms` : "0ms" }}
                    className={cn(
                      "flex items-center gap-3 rounded-xl px-3 py-2.5",
                      "transition-all duration-300",
                      sidebarVisible ? "opacity-100 translate-x-0" : "opacity-0 -translate-x-4"
                    )}
                  >
                    <Link
                      href={href}
                      onClick={closeSidebar}
                      className={cn(
                        "flex items-center gap-3 w-full rounded-xl px-3 py-2.5 -mx-3 transition-colors",
                        active ? "bg-blue-50 text-blue-600" : "text-gray-600 hover:bg-gray-100"
                      )}
                    >
                      <Icon size={18} className="shrink-0" />
                      <span className="text-sm font-medium">{label}</span>
                    </Link>
                  </div>
                );
              })}
            </aside>
          </div>
        )}

        {/* Desktop sidebar — collapses to icons, expands to show labels */}
        <aside
          className={cn(
            "hidden sm:flex bg-white border-r border-gray-200 flex-col pt-4 pb-6 gap-1 shrink-0 overflow-hidden",
            "transition-[width] duration-300 ease-in-out",
            desktopExpanded ? "w-52" : "w-16 md:w-20"
          )}
        >
          {navItems.map(({ label, icon: Icon, href }) => {
            const active = pathname === href;
            return (
              <div
                key={label}
                className={cn(
                  "flex items-center mx-2 rounded-xl",
                  desktopExpanded ? "px-3 py-2.5 gap-3" : "flex-col justify-center py-2 gap-1"
                )}
              >
                <Link
                  href={href}
                  className={cn(
                    "flex items-center w-full rounded-xl transition-colors",
                    desktopExpanded ? "gap-3 px-3 py-2.5" : "flex-col justify-center py-2 gap-1",
                    active ? "bg-blue-50 text-blue-600" : "text-gray-500 hover:bg-gray-100"
                  )}
                >
                  <div className={cn("flex items-center justify-center shrink-0", !desktopExpanded && "w-10 h-10 rounded-xl", active && !desktopExpanded && "bg-blue-100")}>
                    <Icon size={18} />
                  </div>
                  <span
                    className={cn(
                      "whitespace-nowrap transition-all duration-200",
                      desktopExpanded ? "text-sm font-medium opacity-100" : "text-[10px] opacity-100"
                    )}
                  >
                    {label}
                  </span>
                </Link>
              </div>
            );
          })}
        </aside>

        <main className="flex-1 p-4 md:p-6 overflow-y-auto min-w-0 min-h-0 pb-safe">{children}</main>
      </div>
    </div>
  );
}
