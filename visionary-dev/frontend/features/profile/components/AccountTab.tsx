"use client";

import { useState, useEffect } from "react";
import FloatingInput from "@/components/ui/FloatingInput";
import { SectionCard, PillButton } from "./shared";
import { getUserProfile, updateUserProfile, changePassword, requestAccountDeletion } from "@/lib/auth";
import { useAuthStore } from "@/store/authStore";
import { useRouter } from "next/navigation";
import type { UserProfile } from "@/types";

export default function AccountTab() {
  const clear = useAuthStore((s) => s.clear);
  const router = useRouter();

  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);

  // Editable fields
  const [fullName, setFullName] = useState("");
  const [schoolName, setSchoolName] = useState("");
  const [phone, setPhone] = useState("");

  // Saving state
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [saveSuccess, setSaveSuccess] = useState(false);

  // Delete account
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleteReason, setDeleteReason] = useState("");
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteError, setDeleteError] = useState("");

  // Password form
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [retypePassword, setRetypePassword] = useState("");
  const [pwError, setPwError] = useState("");
  const [pwSaving, setPwSaving] = useState(false);
  const [pwSuccess, setPwSuccess] = useState(false);

  useEffect(() => {
    getUserProfile()
      .then((data) => {
        setProfile(data);
        setFullName(data.fullName ?? "");
        setSchoolName(data.schoolName ?? "");
        setPhone(data.phone ?? "");
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  async function handleUpdateBasic(e: { preventDefault: () => void }) {
    e.preventDefault();
    setSaving(true);
    setSaveError("");
    setSaveSuccess(false);
    try {
      const updated = await updateUserProfile({ fullName, schoolName, phone });
      setProfile(updated);
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch {
      setSaveError("Failed to update. Please try again.");
    } finally {
      setSaving(false);
    }
  }

  async function handleChangePassword(e: { preventDefault: () => void }) {
    e.preventDefault();
    if (newPassword !== retypePassword) {
      setPwError("Passwords do not match.");
      return;
    }
    if (newPassword.length < 6) {
      setPwError("New password must be at least 6 characters.");
      return;
    }
    setPwError("");
    setPwSaving(true);
    setPwSuccess(false);
    try {
      await changePassword(currentPassword, newPassword);
      setPwSuccess(true);
      setCurrentPassword("");
      setNewPassword("");
      setRetypePassword("");
      setTimeout(() => setPwSuccess(false), 3000);
    } catch (err: any) {
      setPwError(err?.response?.data?.message ?? "Failed to change password.");
    } finally {
      setPwSaving(false);
    }
  }

  return (
    <div className="flex flex-col gap-5">
      {/* ── Basic Details ── */}
      <SectionCard>
        <h2 className="text-lg font-semibold text-gray-900 mb-6">Basic Details</h2>

        {loading ? (
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-12 bg-gray-100 rounded-lg animate-pulse" />
            ))}
          </div>
        ) : (
          <form onSubmit={handleUpdateBasic}>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
              {/* Editable */}
              <FloatingInput
                label="Full name"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
              />
              {/* Email — readonly */}
              <FloatingInput
                label="Email address"
                type="email"
                value={profile?.email ?? ""}
                readOnly
                className="bg-gray-50 cursor-not-allowed"
              />
              {/* Editable */}
              <FloatingInput
                label="School name"
                value={schoolName}
                onChange={(e) => setSchoolName(e.target.value)}
              />
              {/* Board — readonly, show name */}
              <FloatingInput
                label="Board"
                value={profile?.boardName ?? ""}
                readOnly
                className="bg-gray-50 cursor-not-allowed"
              />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
              {/* Grade — readonly, show name */}
              <FloatingInput
                label="Grade"
                value={profile?.gradeName ?? ""}
                readOnly
                className="bg-gray-50 cursor-not-allowed"
              />
              {/* Phone with dial-code prefix — editable */}
              <div className="flex gap-2">
                <div className="relative shrink-0">
                  <div
                    className="h-full flex items-center px-4 border border-gray-300 rounded-lg text-sm text-gray-700 bg-white"
                    style={{ paddingTop: 21, paddingBottom: 21 }}
                  >
                    +91
                  </div>
                </div>
                <FloatingInput
                  label="Phone"
                  type="tel"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  className="flex-1"
                />
              </div>
            </div>

            {saveError && (
              <p className="text-sm text-red-500 mb-3">{saveError}</p>
            )}
            {saveSuccess && (
              <p className="text-sm text-green-600 mb-3">Profile updated successfully.</p>
            )}

            <PillButton type="submit" disabled={saving}>
              {saving ? "Saving..." : "Update"}
            </PillButton>
          </form>
        )}
      </SectionCard>

      {/* ── Password ── */}
      <SectionCard>
        <h2 className="text-lg font-semibold text-gray-900 mb-6">Password</h2>
        <form onSubmit={handleChangePassword}>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
            <FloatingInput
              label="Current password"
              type="password"
              value={currentPassword}
              onChange={(e) => {
                setCurrentPassword(e.target.value);
                setPwError("");
              }}
            />
            <FloatingInput
              label="New password"
              type="password"
              value={newPassword}
              onChange={(e) => {
                setNewPassword(e.target.value);
                setPwError("");
              }}
            />
            <FloatingInput
              label="Retype password"
              type="password"
              value={retypePassword}
              onChange={(e) => {
                setRetypePassword(e.target.value);
                setPwError("");
              }}
              error={pwError}
            />
          </div>
          {pwError && <p className="text-sm text-red-500 mb-3">{pwError}</p>}
          {pwSuccess && <p className="text-sm text-green-600 mb-3">Password changed successfully.</p>}
          <PillButton type="submit" disabled={pwSaving}>
            {pwSaving ? "Saving..." : "Change password"}
          </PillButton>
        </form>
      </SectionCard>

      {/* ── Delete Account ── */}
      <SectionCard>
        <h2 className="text-lg font-semibold text-gray-900 mb-3">Delete Account</h2>
        <p className="text-sm text-gray-500 leading-relaxed mb-4 max-w-xl">
          If you delete your account, your personal information will be wiped from Visionary&apos;s
          servers, all of your course activity will be anonymized and any certificates earned will be
          deleted. This action cannot be undone! Cancel any active subscriptions before you delete
          your account.
        </p>

        {!showDeleteConfirm ? (
          <PillButton variant="danger" onClick={() => setShowDeleteConfirm(true)}>
            Delete your account
          </PillButton>
        ) : (
          <div className="max-w-xl space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Reason for leaving <span className="text-gray-400 font-normal">(optional)</span>
              </label>
              <textarea
                value={deleteReason}
                onChange={(e) => setDeleteReason(e.target.value)}
                rows={3}
                placeholder="Tell us why you're leaving..."
                className="w-full border border-gray-300 rounded-xl px-4 py-3 text-sm text-gray-700 resize-none focus:outline-none focus:ring-2 focus:ring-red-300"
              />
            </div>

            {deleteError && (
              <p className="text-sm text-red-500">{deleteError}</p>
            )}

            <div className="flex items-center gap-3">
              <PillButton
                variant="danger"
                disabled={deleteLoading}
                onClick={async () => {
                  setDeleteLoading(true);
                  setDeleteError("");
                  try {
                    await requestAccountDeletion(deleteReason || undefined);
                    clear();
                    router.push("/login?deleted=1");
                  } catch (err: any) {
                    setDeleteError(err?.response?.data?.message ?? "Failed to submit request.");
                    setDeleteLoading(false);
                  }
                }}
              >
                {deleteLoading ? "Submitting..." : "Confirm deletion"}
              </PillButton>
              <PillButton onClick={() => { setShowDeleteConfirm(false); setDeleteReason(""); setDeleteError(""); }}>
                Cancel
              </PillButton>
            </div>
          </div>
        )}
      </SectionCard>
    </div>
  );
}
