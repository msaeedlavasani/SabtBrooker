"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import api from "@/lib/api";
import { 
  ArrowRight, 
  MapPin, 
  CheckCircle2, 
  Loader2,
  FileText,
  Map as MapIcon,
  AlertCircle
} from "lucide-react";
import Link from "next/link";

export default function ExpertCaseDetailPage() {
  const { id } = useParams();
  const router = useRouter();
  const [c, setCase] = useState<any>(null);
  const [mapService, setMapService] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const fetchData = async () => {
    try {
      const caseRes = await api.get(`/v1/cases/${id}`);
      setCase(caseRes.data);
      const mapRes = await api.get(`/v1/map-services/${id}`);
      setMapService(mapRes.data);
    } catch (err) {
      router.push("/expert/dashboard");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [id]);

  const handleStartFieldwork = async () => {
    setSubmitting(true);
    try {
      await api.post(`/v1/map-services/${id}/fieldwork/start`);
      await fetchData();
    } catch (err: any) {
      alert(err.response?.data?.error || "خطا در شروع عملیات میدانی");
    } finally {
      setSubmitting(false);
    }
  };

  const handleFinalApprove = async () => {
    setSubmitting(true);
    try {
      await api.post(`/v1/map-services/${id}/submit`);
      alert("نقشه با موفقیت تایید و به سازمان ثبت ارسال شد");
      router.push("/expert/dashboard");
    } catch (err: any) {
      alert(err.response?.data?.error || "خطا در تایید نهایی");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <div className="flex min-h-screen items-center justify-center text-navy font-bold">در حال بارگذاری اطلاعات کارشناسی...</div>;

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <header className="bg-navy p-5 text-white shadow-lg border-b-4 border-brass">
        <div className="mx-auto flex max-w-4xl items-center justify-between">
          <div className="flex items-center gap-4">
            <Link href="/expert/dashboard" className="rounded-full bg-white/10 p-2 hover:bg-white/20">
              <ArrowRight size={20} />
            </Link>
            <h1 className="text-xl font-bold">جزئیات کارشناسی پرونده</h1>
          </div>
          <span className="font-mono text-xs opacity-60">MAP-ID: {mapService?.id}</span>
        </div>
      </header>

      <main className="mx-auto max-w-4xl p-6 space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          
          {/* Case Info */}
          <div className="md:col-span-2 space-y-6">
            <section className="bg-white rounded-2xl border border-line p-6 shadow-sm">
              <h3 className="mb-4 flex items-center gap-2 font-bold text-navy">
                <MapPin size={20} className="text-brass" />
                موقعیت ملک مورد کارشناسی
              </h3>
              <div className="grid grid-cols-2 gap-4 text-sm bg-gray-50 p-4 rounded-xl">
                <div>
                  <span className="text-[11px] text-ink-soft">استان / شهر</span>
                  <div className="font-bold text-navy">{c.province}، {c.city}</div>
                </div>
                <div>
                  <span className="text-[11px] text-ink-soft">نشانی</span>
                  <div className="font-bold text-navy text-xs">{c.address_detail}</div>
                </div>
              </div>
            </section>

            <section className="bg-white rounded-2xl border border-line p-6 shadow-sm">
              <h3 className="mb-6 flex items-center gap-2 font-bold text-navy">
                <MapIcon size={20} className="text-success" />
                وضعیت فرآیند و مدارک
              </h3>
              
              <div className="space-y-6">
                {/* Step 1: Consent */}
                <div className="flex gap-4">
                  <div className={`h-8 w-8 flex-shrink-0 flex items-center justify-center rounded-full ${mapService.consent_granted_at ? "bg-success text-white" : "bg-gray-100 text-ink-soft"}`}>
                    {mapService.consent_granted_at ? <CheckCircle2 size={16} /> : "۱"}
                  </div>
                  <div className="flex-1">
                    <div className="text-sm font-bold text-navy">اخذ رضایت از متقاضی</div>
                    <p className="text-[11px] text-ink-soft mt-1">
                      {mapService.consent_granted_at ? `در تاریخ ${new Date(mapService.consent_granted_at).toLocaleDateString('fa-IR')} تایید شد.` : "در انتظار تایید OTP توسط متقاضی..."}
                    </p>
                  </div>
                </div>

                {/* Step 2: Fieldwork */}
                <div className="flex gap-4">
                  <div className={`h-8 w-8 flex-shrink-0 flex items-center justify-center rounded-full ${["fieldwork_done", "submitted_to_org", "approved"].includes(mapService.status) ? "bg-success text-white" : (mapService.status === "fieldwork_in_progress" ? "bg-brass text-white animate-pulse" : "bg-gray-100 text-ink-soft")}`}>
                    {["fieldwork_done", "submitted_to_org", "approved"].includes(mapService.status) ? <CheckCircle2 size={16} /> : "۲"}
                  </div>
                  <div className="flex-1">
                    <div className="text-sm font-bold text-navy">عملیات میدانی و تهیه نقشه</div>
                    {mapService.status === "expert_assigned" && mapService.consent_granted_at && (
                      <button 
                        onClick={handleStartFieldwork}
                        disabled={submitting}
                        className="mt-3 bg-navy text-white px-4 py-2 rounded-lg text-xs font-bold hover:bg-navy-light disabled:opacity-50"
                      >
                        {submitting ? <Loader2 className="animate-spin" /> : "تایید حضور و شروع عملیات میدانی"}
                      </button>
                    )}
                    {mapService.status === "fieldwork_in_progress" && (
                      <div className="mt-3 bg-brass/10 p-3 rounded-lg border border-brass/20 text-xs text-brass-ink font-bold">
                        کارشناس گرامی، لطفاً پس از بازدید و تهیه نقشه، مدارک را در همین سامانه بارگذاری نمایید.
                      </div>
                    )}
                  </div>
                </div>

                {/* Step 3: Final Approval */}
                {mapService.status === "fieldwork_done" && (
                  <div className="mt-10 pt-6 border-t border-line text-center">
                    <div className="mb-4 inline-flex items-center gap-2 bg-success/10 text-success px-4 py-2 rounded-full text-xs font-bold">
                      <FileText size={16} />
                      مدارک و نقشه بارگذاری شده است
                    </div>
                    <button 
                      onClick={handleFinalApprove}
                      disabled={submitting}
                      className="w-full bg-success text-white p-4 rounded-xl font-bold shadow-lg hover:scale-[1.01] transition-transform disabled:opacity-50"
                    >
                      {submitting ? <Loader2 className="animate-spin" /> : "تایید نهایی و ارسال به سازمان ثبت"}
                    </button>
                  </div>
                )}
              </div>
            </section>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            <div className="bg-white rounded-2xl border border-line p-6 shadow-sm">
              <h3 className="mb-4 text-sm font-bold text-navy">اطلاعات فنی</h3>
              <div className="space-y-3 text-[11px]">
                <div className="flex justify-between border-b border-gray-50 pb-2">
                  <span className="text-ink-soft">نوع ملک:</span>
                  <span className="font-bold text-navy">{mapService.property_type || "نامشخص"}</span>
                </div>
                <div className="flex justify-between border-b border-gray-50 pb-2">
                  <span className="text-ink-soft">مساحت تقریبی:</span>
                  <span className="font-bold text-navy font-mono">{mapService.approx_area_sqm || "۰"} m²</span>
                </div>
              </div>
            </div>

            <div className="bg-navy p-6 rounded-2xl text-white shadow-xl">
              <h4 className="font-bold text-brass-light mb-2 flex items-center gap-2">
                <AlertCircle size={16} />
                توجه قانونی
              </h4>
              <p className="text-[10px] leading-relaxed opacity-80">
                هرگونه ثبت اطلاعات خلاف واقع توسط کارشناس، موجب مسئولیت‌های قانونی مقرر در ماده ۱۱۴ آیین‌نامه اجرایی خواهد بود.
              </p>
            </div>
          </div>

        </div>
      </main>
    </div>
  );
}
