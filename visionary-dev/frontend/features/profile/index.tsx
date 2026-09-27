"use client";

import { useState } from "react";
import Breadcrumb from "@/components/ui/Breadcrumb";
import Tabs from "@/components/ui/Tabs";
import AccountTab from "./components/AccountTab";
import PreferencesTab from "./components/PreferencesTab";

type Tab = "account" | "preferences";

const TABS: { id: Tab; label: string }[] = [
    { id: "account", label: "Account" },
    { id: "preferences", label: "Preferences" },
];

export default function ProfilePage() {
    const [activeTab, setActiveTab] = useState<Tab>("account");

    return (
        <div className="w-full min-w-0 -m-4 md:-m-6 p-4 md:p-6 min-h-full" style={{ backgroundColor: "#F0F0F0" }}>
            {/* Breadcrumb */}
            <Breadcrumb items={[{ label: "Account settings" }]} className="mb-6" />

            <div className="flex flex-col sm:flex-row gap-5 items-start">
                {/* ── Left sidebar tabs ── */}
                <div className="flex sm:flex-col shrink-0 w-full sm:w-64"
                    style={{ borderRadius: 24, padding: 32, gap: 28, minHeight: 192, backgroundColor: "#F0F0F0", border: "1px solid #C7C7C7" }}
                >
                    <Tabs
                        variant="pill"
                        tabs={TABS.map((t) => t.label)}
                        activeTab={TABS.find((t) => t.id === activeTab)?.label ?? ""}
                        onChange={(label) => {
                            const found = TABS.find((t) => t.label === label);
                            if (found) setActiveTab(found.id);
                        }}
                    />
                </div>

                {/* ── Right content ── */}
                <div className="flex-1 min-w-0">
                    {activeTab === "account" ? <AccountTab /> : <PreferencesTab />}
                </div>
            </div>
        </div>
    );
}
