"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import api from "@/lib/api";
import { 
  ArrowRight, 
  ArrowLeft, 
  MapPin, 
  User, 
  Building2, 
  CheckCircle2, 
  Loader2,
  ShieldCheck
} from "lucide-react";
import Link from "next/link";

const STEPS = [
  { id: 1, title: "هویت متقاضی", icon: <User size={18} /> },
  { id: 2, title: "مشخصات ملک", icon: <MapPin size={18} /> },
  { id: 3, title: "تایید نهایی", icon: <CheckCircle2 size={18} /> },
];

export default function NewCasePage() {
  const router = useRouter();
  const [currentStep, setCurrentStep] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  
  const [formData, setFormData] = useState({
    province: "",
    city: "",
    district: "",
    village: "",
    address_detail: "",
    capacity: "principal",
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async () => {
    setLoading(true);
    setError("");
    try {
      // 1. Create Case
      const res = await api.post("/v1/cases", {
        province: formData.province,
        city: formData.city,
        district: formData.district || null,
        village: formData.village || null,
        address_detail: formData.address_detail || null,
      });
      
      const caseId = res.data.id;
      
      // 2. Update Capacity (if needed, but for now we use the endpoint defined in handoff)
      if (formData.capacity !== "principal") {
        await api.put(`/v1/cases/${caseId}/capacity`, { capacity: formData.capacity });
      }

      router.push("/dashboard");
    } catch (err: any) {
      setError(err.response?.data?.error || "خطا در ثبت پرونده. لطفا اطلاعات را بررسی کنید.");
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-paper pb-20">
      <header className="bg-navy p-5 text-white shadow-lg">
        <div className="mx-auto flex max-w-4xl items-center justify-between">
          <div className="flex items-center gap-4">
            <Link href="/dashboard" className="rounded-full bg-white/10 p-2 hover:bg-white/20">
              <ArrowRight size={20} />
            </Link>
            <h1 className="text-xl font-bold">تشکیل پرونده جدید</h1>
          </div>
          <div className="flex items-center gap-2 text-xs text-brass-light">
            <ShieldCheck size={16} />
            امنیت متصل به سامانه ثنا
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-3xl p-6">
        {/* Step Indicator */}
        <div className="mb-8 flex items-center justify-between px-4">
          {STEPS.map((step) => (
            <div key={step.id} className="flex flex-col items-center gap-2">
              <div className={`flex h-10 w-10 items-center justify-center rounded-full border-2 transition-colors ${
                currentStep >= step.id ? "border-navy bg-navy text-white" : "border-line bg-white text-ink-soft"
              }`}>
                {step.icon}
              </div>
              <span className={`text-[11px] font-bold ${currentStep >= step.id ? "text-navy" : "text-ink-soft"}`}>
                {step.title}
              </span>
            </div>
          ))}
          <div className="absolute left-1/2 -z-10 h-[2px] w-2/3 -translate-x-1/2 bg-line md:w-1/2" />
        </div>

        {error && (
          <div className="mb-6 rounded-lg border border-error bg-error-bg p-4 text-sm text-error">
            {error}
          </div>
        )}

        <div className="rounded-2xl border border-line bg-white p-8 shadow-sm">
          {currentStep === 1 && (
            <div className="space-y-6">
              <h2 className="text-lg font-bold text-navy">هویت و سمت متقاضی</h2>
              <p className="text-xs text-ink-soft">لطفاً ظرفیت خود را در این پرونده مشخص کنید.</p>
              
              <div className="space-y-4">
                <div className="field">
                  <label className="text-sm font-semibold text-ink">سمت متقاضی در پرونده</label>
                  <select 
                    name="capacity" 
                    value={formData.capacity} 
                    onChange={handleChange}
                    className="mt-1 block w-full rounded-lg border border-line bg-gray-50 p-3 text-sm focus:border-navy focus:bg-white focus:outline-none"
                  >
                    <option value="principal">اصیل (مالک)</option>
                    <option value="legal_rep_natural">نماینده قانونی شخص حقیقی (وکیل/قیم)</option>
                    <option value="legal_rep_legal">نماینده قانونی شخص حقوقی (مدیرعامل)</option>
                  </select>
                </div>

                <div className="rounded-lg bg-brass/5 p-4 text-[12px] leading-relaxed text-brass-ink border border-brass/20">
                  توجه: در صورت انتخاب هر گزینه‌ای غیر از "اصیل"، در مراحل بعد باید مدارک مثبته (وکالتنامه، آگهی تغییرات و ...) را بارگذاری نمایید.
                </div>
              </div>

              <div className="mt-8 flex justify-end">
                <button 
                  onClick={() => setCurrentStep(2)}
                  className="flex items-center gap-2 rounded-lg bg-navy px-6 py-3 font-bold text-white hover:opacity-90"
                >
                  مرحله بعد
                  <ArrowLeft size={18} />
                </button>
              </div>
            </div>
          )}

          {currentStep === 2 && (
            <div className="space-y-6">
              <h2 className="text-lg font-bold text-navy">موقعیت مکانی و مشخصات ملک</h2>
              
              <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
                <div className="field">
                  <label className="text-sm font-semibold text-ink">استان</label>
                  <input 
                    name="province" 
                    value={formData.province} 
                    onChange={handleChange}
                    placeholder="مثلاً: تهران"
                    className="mt-1 block w-full rounded-lg border border-line bg-gray-50 p-3 text-sm focus:border-navy focus:bg-white focus:outline-none"
                  />
                </div>
                <div className="field">
                  <label className="text-sm font-semibold text-ink">شهرستان / شهر</label>
                  <input 
                    name="city" 
                    value={formData.city} 
                    onChange={handleChange}
                    placeholder="مثلاً: تهران"
                    className="mt-1 block w-full rounded-lg border border-line bg-gray-50 p-3 text-sm focus:border-navy focus:bg-white focus:outline-none"
                  />
                </div>
                <div className="field">
                  <label className="text-sm font-semibold text-ink">بخش / دهستان (اختیاری)</label>
                  <input 
                    name="district" 
                    value={formData.district} 
                    onChange={handleChange}
                    className="mt-1 block w-full rounded-lg border border-line bg-gray-50 p-3 text-sm focus:border-navy focus:bg-white focus:outline-none"
                  />
                </div>
                <div className="field">
                  <label className="text-sm font-semibold text-ink">روستا (اختیاری)</label>
                  <input 
                    name="village" 
                    value={formData.village} 
                    onChange={handleChange}
                    className="mt-1 block w-full rounded-lg border border-line bg-gray-50 p-3 text-sm focus:border-navy focus:bg-white focus:outline-none"
                  />
                </div>
              </div>

              <div className="field">
                <label className="text-sm font-semibold text-ink">نشانی دقیق پستی</label>
                <textarea 
                  name="address_detail" 
                  value={formData.address_detail} 
                  onChange={handleChange}
                  rows={3}
                  className="mt-1 block w-full rounded-lg border border-line bg-gray-50 p-3 text-sm focus:border-navy focus:bg-white focus:outline-none"
                />
              </div>

              <div className="mt-8 flex justify-between">
                <button 
                  onClick={() => setCurrentStep(1)}
                  className="flex items-center gap-2 rounded-lg border border-line px-6 py-3 font-bold text-ink-soft hover:bg-gray-50"
                >
                  <ArrowRight size={18} />
                  قبلی
                </button>
                <button 
                  onClick={() => setCurrentStep(3)}
                  disabled={!formData.province || !formData.city || !formData.address_detail}
                  className="flex items-center gap-2 rounded-lg bg-navy px-6 py-3 font-bold text-white hover:opacity-90 disabled:opacity-50"
                >
                  مرحله بعد
                  <ArrowLeft size={18} />
                </button>
              </div>
            </div>
          )}

          {currentStep === 3 && (
            <div className="space-y-6">
              <h2 className="text-lg font-bold text-navy">تایید و ثبت نهایی</h2>
              
              <div className="rounded-xl border border-line bg-gray-50 p-6 space-y-4">
                <SummaryItem label="استان و شهر" value={`${formData.province}، ${formData.city}`} />
                <SummaryItem label="نشانی" value={formData.address_detail} />
                <SummaryItem label="سمت" value={formData.capacity === "principal" ? "اصیل" : "نماینده"} />
              </div>

              <div className="rounded-lg bg-navy/5 p-4 text-[12px] leading-relaxed text-navy border border-navy/10 flex gap-3">
                <ShieldCheck className="text-navy flex-shrink-0" />
                <p>با کلیک روی دکمه ثبت، شما صحت اطلاعات فوق را تایید نموده و درخواست خود را جهت شروع فرآیند کارگزاری ماده ۱۰ ارسال می‌کنید.</p>
              </div>

              <div className="mt-8 flex justify-between">
                <button 
                  onClick={() => setCurrentStep(2)}
                  className="flex items-center gap-2 rounded-lg border border-line px-6 py-3 font-bold text-ink-soft hover:bg-gray-50"
                >
                  <ArrowRight size={18} />
                  ویرایش
                </button>
                <button 
                  onClick={handleSubmit}
                  disabled={loading}
                  className="flex items-center gap-2 rounded-lg bg-navy px-8 py-3 font-bold text-white shadow-lg transition-transform hover:scale-[1.02] disabled:opacity-50"
                >
                  {loading ? <Loader2 className="animate-spin" /> : "ثبت نهایی و تشکیل پرونده"}
                </button>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

function SummaryItem({ label, value }: { label: string, value: string }) {
  return (
    <div className="flex justify-between border-b border-line pb-2 text-sm">
      <span className="text-ink-soft font-semibold">{label}</span>
      <span className="text-navy font-bold">{value}</span>
    </div>
  );
}
