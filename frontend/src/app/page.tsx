"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import api from "@/lib/api";
import { Loader2, Phone, ShieldCheck } from "lucide-react";

export default function LoginPage() {
  const [mobile, setMobile] = useState("");
  const [otp, setOtp] = useState("");
  const [step, setStep] = useState(1); // 1: Mobile, 2: OTP
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [devOtp, setDevOtp] = useState("");
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

  const handleVerifyOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await api.post("/v1/auth/otp/verify", { mobile, otp });
      localStorage.setItem("access_token", res.data.access_token);
      
      // Get user profile to determine role
      const userRes = await api.get("/v1/auth/me");
      if (userRes.data.role === "expert") {
        router.push("/expert/dashboard");
      } else {
        router.push("/dashboard");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "کد تایید نامعتبر است");
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
            ورود به پنل کاربری با شماره همراه
          </p>
        </div>

        {error && (
          <div className="rounded-lg border border-error bg-error-bg p-3 text-sm text-error">
            {error}
          </div>
        )}

        {step === 1 ? (
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
                شماره باید به نام متقاضی در سامانه شاهکار ثبت شده باشد.
              </p>
            </div>

            <button
              disabled={loading}
              className="flex w-full items-center justify-center rounded-lg bg-navy p-3 font-bold text-white transition-opacity hover:opacity-90 disabled:opacity-50"
            >
              {loading ? <Loader2 className="animate-spin" /> : "ورود مستقیم به پنل"}
            </button>
          </form>
        ) : (
          <form className="mt-8 space-y-6" onSubmit={handleVerifyOtp}>
            <div className="space-y-2">
              <label className="text-sm font-semibold text-ink text-center block">
                کد تایید ۵ رقمی را وارد کنید
              </label>
              <input
                type="text"
                required
                maxLength={5}
                className="block w-full rounded-lg border border-line bg-gray-50 p-4 text-center text-2xl font-bold tracking-[1rem] focus:border-navy focus:bg-white focus:outline-none"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
              />
              {devOtp && (
                <div className="text-center">
                  <span className="text-[11px] text-brass">
                    محیط توسعه: کد تایید {devOtp} است
                  </span>
                </div>
              )}
            </div>

            <div className="flex flex-col gap-3">
              <button
                disabled={loading}
                className="flex w-full items-center justify-center rounded-lg bg-navy p-3 font-bold text-white transition-opacity hover:opacity-90 disabled:opacity-50"
              >
                {loading ? <Loader2 className="animate-spin" /> : "تایید و ورود"}
              </button>
              <button
                type="button"
                className="text-sm text-ink-soft hover:text-navy"
                onClick={() => setStep(1)}
              >
                ویرایش شماره موبایل
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
