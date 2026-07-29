"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import api from "@/lib/api";
import { 
  ArrowRight, 
  MapPin, 
  Clock, 
  CheckCircle2, 
  AlertCircle,
  FileText,
  Map as MapIcon,
  ChevronLeft,
  Loader2,
  ExternalLink
} from "lucide-react";
import Link from "next/link";

export default function CaseDetailPage() {
  const { id } = useParams();
  const router = useRouter();
  const [c, setCase] = useState<any>(null);
  const [mapService, setMapService] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const fetchData = async () => {
    try {
      const caseRes = await api.get(`/v1/cases/${id}`);
      setCase(caseRes.data);
      
      if (caseRes.data.status !== "draft") {
        const mapRes = await api.get(`/v1/map-services/${id}`);
        setMapService(mapRes.data);
      }
    } catch (err) {
      setError("خطا در بارگذاری اطلاعات پرونده");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [id]);

  const handleSubmitToMap = async () => {
    setSubmitting(true);
    try {
      await api.post(`/v1/cases/${id}/submit`);
      await fetchData();
    } catch (err: any) {
      alert(err.response?.data?.error || "خطا در شروع فرآیند نقشه");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <div className="flex min-h-screen items-center justify-center text-navy font-bold">در حال بارگذاری...</div>;
  if (!c) return <div className="flex min-h-screen items-center justify-center text-error font-bold">پرونده یافت نشد</div>;

  return (
    <div className="min-h-screen bg-paper pb-20">
      <header className="bg-navy p-5 text-white shadow-lg">
        <div className="mx-auto flex max-w-4xl items-center justify-between">
          <div className="flex items-center gap-4">
            <Link href="/dashboard" className="rounded-full bg-white/10 p-2 hover:bg-white/20">
              <ArrowRight size={20} />
            </Link>
            <h1 className="text-xl font-bold">جزئیات پرونده</h1>
          </div>
          <span className="font-mono text-xs opacity-60">{c.id}</span>
        </div>
      </header>

      <main className="mx-auto max-w-4xl p-6 space-y-6">
        {/* Status Tracker */}
        <div className="rounded-2xl border border-line bg-white p-6 shadow-sm">
          <h2 className="mb-6 text-sm font-bold text-ink-soft">چرخه حیات پرونده</h2>
          <div className="flex items-center justify-between relative">
            <StatusNode active={true} done={true} label="تشکیل پرونده" />
            <StatusNode active={c.status !== "draft"} done={c.status !== "draft" && c.status !== "map_in_progress"} label="نقشه‌برداری" />
            <StatusNode active={c.status === "claim_in_progress"} done={false} label="درج ادعا" />
            <StatusNode active={c.status === "cert_in_progress"} done={false} label="گواهی اقدام" />
            
            <div className="absolute left-0 right-0 top-5 -z-10 h-[2px] bg-line" />
          </div>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          <div className="md:col-span-2 space-y-6">
            <section className="rounded-2xl border border-line bg-white p-6 shadow-sm">
              <h3 className="mb-4 flex items-center gap-2 font-bold text-navy">
                <MapPin size={20} className="text-brass" />
                مشخصات ملک
              </h3>
              <div className="grid grid-cols-2 gap-y-4 text-sm">
                <InfoItem label="استان و شهر" value={`${c.province}، ${c.city}`} />
                <InfoItem label="بخش/روستا" value={`${c.district || "-"} / ${c.village || "-"}`} />
                <div className="col-span-2 mt-2 pt-2 border-t border-gray-50">
                  <span className="text-xs text-ink-soft">نشانی دقیق:</span>
                  <p className="mt-1 font-semibold text-navy">{c.address_detail}</p>
                </div>
              </div>
            </section>

            {c.status === "draft" ? (
              <section className="rounded-2xl border-2 border-dashed border-brass/30 bg-brass/5 p-8 text-center">
                <MapIcon size={48} className="mx-auto mb-4 text-brass opacity-40" />
                <h3 className="text-lg font-bold text-navy">شروع فرآیند نقشه‌برداری</h3>
                <p className="mt-2 text-sm text-ink-soft">
                  پرونده شما در حالت پیش‌نویس است. برای ادامه زنجیره، باید درخواست نقشه‌برداری را ارسال کنید.
                </p>
                <button 
                  onClick={handleSubmitToMap}
                  disabled={submitting}
                  className="mt-6 flex w-full items-center justify-center gap-2 rounded-xl bg-navy p-4 font-bold text-white shadow-md hover:scale-[1.01] transition-transform disabled:opacity-50"
                >
                  {submitting ? <Loader2 className="animate-spin" /> : <><CheckCircle2 size={20} /> تایید و ارسال جهت نقشه‌برداری</>}
                </button>
              </section>
            ) : (
              <section className="rounded-2xl border border-line bg-white p-6 shadow-sm">
                <div className="flex items-center justify-between mb-6">
                  <h3 className="flex items-center gap-2 font-bold text-navy">
                    <MapIcon size={20} className="text-success" />
                    وضعیت سرویس نقشه
                  </h3>
                  {mapService?.tracking_code && (
                    <span className="font-mono text-xs font-bold text-success bg-success-bg px-3 py-1 rounded-full border border-success/20">
                      {mapService.tracking_code}
                    </span>
                  )}
                </div>
                
                <div className="space-y-4">
                  <WorkflowStep 
                    label="تخصیص کارشناس" 
                    status={mapService?.status !== "pending_expert_assignment" ? "done" : "current"} 
                  />
                  <WorkflowStep 
                    label="عملیات میدانی" 
                    status={mapService?.status === "fieldwork_in_progress" ? "current" : (["fieldwork_done", "submitted_to_org", "approved"].includes(mapService?.status) ? "done" : "pending")} 
                  />
                  <WorkflowStep 
                    label="صدور کد رهگیری مانا" 
                    status={mapService?.status === "approved" ? "done" : "pending"} 
                  />
                </div>
              </section>
            )}
          </div>

          <div className="space-y-6">
            <section className="rounded-2xl border border-line bg-white p-6 shadow-sm">
              <h3 className="mb-4 flex items-center gap-2 font-bold text-navy text-sm">
                <FileText size={18} />
                اطلاعات متقاضی
              </h3>
              <div className="space-y-3 text-xs">
                <div className="flex justify-between">
                  <span className="text-ink-soft">ظرفیت:</span>
                  <span className="font-bold text-navy">{c.applicant_capacity === "principal" ? "اصیل" : "نماینده"}</span>
                </div>
              </div>
            </section>
          </div>
        </div>
      </main>
    </div>
  );
}

function StatusNode({ active, done, label }: any) {
  return (
    <div className="flex flex-col items-center gap-2 bg-white px-2">
      <div className={`h-10 w-10 flex items-center justify-center rounded-full border-2 transition-colors ${
        done ? "bg-success border-success text-white" : (active ? "bg-navy border-navy text-white" : "bg-white border-line text-ink-soft")
      }`}>
        {done ? <CheckCircle2 size={20} /> : <div className="h-2 w-2 rounded-full bg-current" />}
      </div>
      <span className={`text-[10px] font-bold ${active || done ? "text-navy" : "text-ink-soft"}`}>{label}</span>
    </div>
  );
}

function InfoItem({ label, value }: any) {
  return (
    <div className="space-y-1">
      <span className="text-[11px] text-ink-soft">{label}</span>
      <div className="font-bold text-navy">{value}</div>
    </div>
  );
}

function WorkflowStep({ label, status }: { label: string, status: "done" | "current" | "pending" }) {
  return (
    <div className="flex items-center gap-4">
      <div className={`h-8 w-8 flex items-center justify-center rounded-full border ${
        status === "done" ? "bg-success/10 border-success text-success" : (status === "current" ? "bg-brass/10 border-brass text-brass animate-pulse" : "bg-gray-50 border-line text-ink-soft")
      }`}>
        {status === "done" ? <CheckCircle2 size={16} /> : <div className="h-1.5 w-1.5 rounded-full bg-current" />}
      </div>
      <span className={`text-sm font-bold ${status === "pending" ? "text-ink-soft" : "text-navy"}`}>{label}</span>
    </div>
  );
}
