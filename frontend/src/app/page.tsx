"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import api from "@/lib/api";
import { Loader2, Phone, ShieldCheck } from "lucide-react";

export default function LoginPage() {
  const [mobile, setMobile] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();

  const handleSendOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      // BYPASS SENDING OTP - Just trigger verification with fake code
      const res = await api.post("/v1/auth/otp/verify", { mobile, otp: "1234" });
      localStorage.setItem("access_token", res.data.access_token);
      
      const userRes = await api.get("/v1/auth/me");
      if (userRes.data.role === "expert") {
        router.push("/expert/dashboard");
      } else {
        router.push("/dashboard");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "خطا در ورود به سامانه");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-paper p-4">
      <div className="w-full max-w-md space-y-8 rounded-2xl border border-line bg-white p-8 shadow-sm">
        <div className="text-center">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-navy text-brass-light">
            <ShieldCheck size={32} />
          </div>
          <h2 className="mt-6 text-2xl font-extrabold text-navy">
            سامانه کارگزاری ماده ۱۰
          </h2>
          <p className="mt-2 text-sm text-ink-soft">
            ورود سریع به پنل کاربری (نسخه دمو)
          </p>
        </div>

        {error && (
          <div className="rounded-lg border border-error bg-error-bg p-3 text-sm text-error">
            {error}
          </div>
        )}

        <form className="mt-8 space-y-6" onSubmit={handleSendOtp}>
          <div className="space-y-2">
            <label className="text-sm font-semibold text-ink">
              شماره تلفن همراه
            </label>
            <div className="relative">
              <input
                type="text"
                required
                placeholder="0912xxxxxxx"
                className="block w-full rounded-lg border border-line bg-gray-50 p-3 pl-10 text-center font-mono focus:border-navy focus:bg-white focus:outline-none"
                value={mobile}
                onChange={(e) => setMobile(e.target.value)}
              />
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 text-ink-soft">
                <Phone size={18} />
              </div>
            </div>
            <p className="text-[11px] text-ink-soft">
              در این نسخه دمو، نیاز به تایید کد اس‌ام‌اس نیست.
            </p>
          </div>

          <button
            disabled={loading}
            className="flex w-full items-center justify-center rounded-lg bg-navy p-3 font-bold text-white transition-opacity hover:opacity-90 disabled:opacity-50"
          >
            {loading ? <Loader2 className="animate-spin" /> : "ورود مستقیم به پنل"}
          </button>
        </form>
      </div>
    </div>
  );
}
