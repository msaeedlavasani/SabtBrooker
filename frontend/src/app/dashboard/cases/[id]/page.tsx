
"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import clsx from "clsx";
import api from "@/lib/api";
import { 
  ArrowRight, 
  MapPin, 
  CheckCircle2, 
  FileText,
  Map as MapIcon,
  Loader2,
  ShieldCheck,
  AlertTriangle,
  History
} from "lucide-react";
import Link from "next/link";
import FormWizard from "@/components/FormWizard";
import { MAP_SCREENS, CLAIM_SCREENS, ACTION_SCREENS, Screen } from "@/lib/workflow-configs";

export default function CaseDetailPage() {
  const { id } = useParams();
  const router = useRouter();
  const [c, setCase] = useState<any>(null);
  const [mapService, setMapService] = useState<any>(null);
  const [claimService, setClaimService] = useState<any>(null);
  const [certService, setCertService] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  
  const fetchData = useCallback(async () => {
    try {
      const caseRes = await api.get(`/v1/cases/${id}`);
      setCase(caseRes.data);
      
      if (caseRes.data.status !== "draft") {
        try {
            const mapRes = await api.get(`/v1/map-services/${id}`);
            setMapService(mapRes.data);
        } catch (e) {}

        if (caseRes.data.status === "claim_in_progress" || caseRes.data.status === "cert_in_progress" || caseRes.data.status === "completed") {
            try {
                const claimRes = await api.get(`/v1/claim-services/${id}`);
                setClaimService(claimRes.data);
            } catch (e) {}
        }

        if (caseRes.data.status === "cert_in_progress" || caseRes.data.status === "completed") {
            try {
                const certRes = await api.get(`/v1/cert-services/${id}`);
                setCertService(certRes.data);
            } catch (e) {}
        }
      }
    } catch (err) {
      setError("خطا در بارگذاری اطلاعات پرونده");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleWorkflowComplete = async (data: any) => {
    setLoading(true);
    try {
        if (c.status === 'draft') {
            // 1. Update Case with additional info from first screens
            await api.patch(`/v1/cases/${id}`, {
                district: data.district,
                village: data.village,
                address_detail: data.address
            });
            // 2. Submit to Map
            await api.post(`/v1/cases/${id}/submit`);
        } else if (c.status === 'map_in_progress' && mapService?.status === 'fieldwork_in_progress') {
            // Submit Fieldwork
            await api.post(`/v1/map-services/${id}/fieldwork/submit`, {
                map_file_id: data.survey_map_file_ref ? "mock-file-id" : "",
                descriptive_table: JSON.parse(data.survey_descriptive_table_json || "{}")
            });
        }
        await fetchData();
    } catch (err: any) {
        alert(err.response?.data?.error || "خطا در عملیات");
    } finally {
        setLoading(false);
    }
  };

  if (loading && !c) return <div className="flex min-h-screen items-center justify-center text-navy font-bold">در حال بارگذاری...</div>;
  if (error) return <div className="flex min-h-screen items-center justify-center text-error font-bold">{error}</div>;
  if (!c) return <div className="flex min-h-screen items-center justify-center text-error font-bold">پرونده یافت نشد</div>;

  const getActiveScreens = (): Screen[] => {
    if (c.status === 'draft') {
        return MAP_SCREENS.filter(s => ['identity', 'location_property'].includes(s.id));
    }
    if (c.status === 'map_in_progress') {
        if (mapService?.status === 'fieldwork_in_progress') {
            return MAP_SCREENS.filter(s => s.id === 'survey');
        }
        if (mapService?.status === 'approved') {
            return MAP_SCREENS.filter(s => s.id === 'final');
        }
        return []; // Waiting states
    }
    if (c.status === 'claim_in_progress') {
        if (claimService?.status === 'expert_assigned') {
            return CLAIM_SCREENS.filter(s => ['identity', 'map_ref', 'expert_registry'].includes(s.id));
        }
        if (claimService?.status === 'documents_verified') {
            return CLAIM_SCREENS.filter(s => ['claim_details', 'consent'].includes(s.id));
        }
        return CLAIM_SCREENS; // Default to all if status unknown or initial
    }
    if (c.status === 'cert_in_progress') {
        return ACTION_SCREENS;
    }
    return [];
  };

  const activeScreens = getActiveScreens();

  return (
    <div className="min-h-screen bg-paper pb-20">
      <header className="bg-navy p-5 text-white shadow-lg border-b border-brass/20">
        <div className="mx-auto flex max-w-5xl items-center justify-between">
          <div className="flex items-center gap-4">
            <Link href="/dashboard" className="rounded-full bg-white/10 p-2 hover:bg-white/20 transition-colors">
              <ArrowRight size={20} />
            </Link>
            <div>
                <h1 className="text-xl font-extrabold tracking-tight">پنل مدیریت پرونده</h1>
                <p className="text-[10px] text-brass-light font-mono opacity-70 uppercase">{c.id}</p>
            </div>
          </div>
          <div className="flex items-center gap-2 bg-white/5 px-4 py-2 rounded-xl border border-white/10">
            <ShieldCheck size={16} className="text-brass-light" />
            <span className="text-[11px] font-bold">اتصال امن کارگزاری</span>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-5xl p-6 space-y-8">
        {/* Progress Tracker (High Fidelity) */}
        <section className="bg-white rounded-2xl border border-line p-8 shadow-sm">
            <div className="flex items-center justify-between relative max-w-3xl mx-auto">
                <StatusStep 
                    icon={<FileText size={18} />} 
                    label="تشکیل پرونده" 
                    status={c.status === 'draft' ? 'current' : 'done'} 
                />
                <div className={clsx("h-[2px] flex-1 mx-2 transition-colors", c.status !== 'draft' ? "bg-navy" : "bg-line")} />
                <StatusStep 
                    icon={<MapIcon size={18} />} 
                    label="نقشه‌برداری" 
                    status={c.status === 'map_in_progress' ? 'current' : (['claim_in_progress', 'cert_in_progress', 'completed'].includes(c.status) ? 'done' : 'pending')} 
                />
                <div className={clsx("h-[2px] flex-1 mx-2 transition-colors", ['claim_in_progress', 'cert_in_progress', 'completed'].includes(c.status) ? "bg-navy" : "bg-line")} />
                <StatusStep 
                    icon={<History size={18} />} 
                    label="درج ادعا" 
                    status={c.status === 'claim_in_progress' ? 'current' : (['cert_in_progress', 'completed'].includes(c.status) ? 'done' : 'pending')} 
                />
                <div className={clsx("h-[2px] flex-1 mx-2 transition-colors", ['cert_in_progress', 'completed'].includes(c.status) ? "bg-navy" : "bg-line")} />
                <StatusStep 
                    icon={<CheckCircle2 size={18} />} 
                    label="سند نهایی" 
                    status={c.status === 'completed' ? 'done' : (c.status === 'cert_in_progress' ? 'current' : 'pending')} 
                />
            </div>
        </section>

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
            {/* Left Sidebar: Summary */}
            <div className="lg:col-span-1 space-y-6">
                <div className="bg-white rounded-2xl border border-line p-6 shadow-sm sticky top-6">
                    <h3 className="text-sm font-extrabold text-navy mb-4 flex items-center gap-2">
                        <MapPin size={18} className="text-brass" />
                        مشخصات پایه
                    </h3>
                    <div className="space-y-4">
                        <SideInfo label="موقعیت" value={`${c.province}، ${c.city}`} />
                        <SideInfo label="وضعیت" value={getStatusLabel(c.status)} color={getStatusColor(c.status)} />
                        <SideInfo label="سمت متقاضی" value={c.applicant_capacity === 'principal' ? 'اصیل' : 'نماینده'} />
                    </div>
                    
                    <div className="mt-8 pt-6 border-t border-line">
                        <div className="bg-navy/5 rounded-xl p-4 border border-navy/10">
                            <h4 className="text-[11px] font-bold text-navy mb-2 flex items-center gap-2">
                                <AlertTriangle size={14} className="text-brass" />
                                راهنمای مرحله
                            </h4>
                            <p className="text-[10px] leading-relaxed text-ink-soft">
                                {getStepGuide(c.status, mapService?.status)}
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            {/* Main Action Area */}
            <div className="lg:col-span-3 space-y-6">
                {activeScreens.length > 0 ? (
                    <div className="bg-white rounded-3xl border border-line p-8 shadow-sm">
                        <FormWizard 
                            screens={activeScreens} 
                            initialData={{
                                applicant_national_code: c.applicant_national_code || "",
                                applicant_phone: c.applicant_phone || "",
                                province: c.province,
                                county: c.city,
                                address: c.address_detail
                            }}
                            onComplete={handleWorkflowComplete}
                            isLoading={loading}
                        />
                    </div>
                ) : (
                    <div className="bg-white rounded-3xl border border-line p-12 text-center shadow-sm">
                        <div className="w-20 h-20 bg-gray-50 rounded-full flex items-center justify-center mx-auto mb-6">
                            <Loader2 className="animate-spin text-navy/20" size={32} />
                        </div>
                        <h3 className="text-xl font-bold text-navy mb-2">در انتظار اقدام سیستم یا کارشناس</h3>
                        <p className="text-sm text-ink-soft max-w-md mx-auto">
                            در حال حاضر پرونده در صف بررسی کارشناسان یا پردازش سیستمی قرار دارد. به محض تغییر وضعیت، مراحل بعدی برای شما فعال خواهد شد.
                        </p>
                        
                        {mapService?.status === 'pending_expert_assignment' && (
                            <div className="mt-8 p-4 bg-brass/5 border border-brass/20 rounded-xl text-brass text-xs font-bold inline-block">
                                وضعیت فعلی: منتظر تخصیص کارشناس امور ثبتی
                            </div>
                        )}
                    </div>
                )}

                {/* Case History / Audit Logs */}
                <div className="bg-white rounded-3xl border border-line p-8 shadow-sm">
                    <h3 className="text-lg font-bold text-navy mb-6 flex items-center gap-2">
                        <History size={20} className="text-navy/40" />
                        تاریخچه فعالیت‌ها
                    </h3>
                    <div className="space-y-4">
                        <HistoryItem 
                            date={new Date(c.created_at).toLocaleDateString('fa-IR')} 
                            time={new Date(c.created_at).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' })}
                            title="تشکیل پرونده اولیه"
                            desc="پرونده توسط متقاضی با موفقیت در سیستم ثبت شد."
                            done
                        />
                        {c.status !== 'draft' && (
                            <HistoryItem 
                                date={new Date(c.updated_at).toLocaleDateString('fa-IR')}
                                time={new Date(c.updated_at).toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit' })}
                                title="ارسال جهت نقشه‌برداری"
                                desc="پرونده جهت طی فرآیند تهیه نقشه ثبتی به کارتابل کارشناسان ارسال شد."
                                done
                            />
                        )}
                    </div>
                </div>
            </div>
        </div>
      </main>
    </div>
  );
}

function StatusStep({ icon, label, status }: { icon: any, label: string, status: 'done' | 'current' | 'pending' }) {
    return (
        <div className="flex flex-col items-center gap-3 group">
            <div className={clsx(
                "w-12 h-12 rounded-full flex items-center justify-center transition-all duration-500 border-2",
                status === 'done' ? "bg-navy border-navy text-white" : 
                (status === 'current' ? "bg-white border-navy text-navy shadow-[0_0_15px_rgba(16,27,51,0.15)] scale-110" : "bg-white border-line text-ink-soft")
            )}>
                {status === 'done' ? <CheckCircle2 size={24} /> : icon}
            </div>
            <span className={clsx(
                "text-[11px] font-extrabold tracking-tight transition-colors",
                status === 'pending' ? "text-ink-soft" : "text-navy"
            )}>
                {label}
            </span>
        </div>
    );
}

function SideInfo({ label, value, color = "text-navy" }: { label: string, value: string, color?: string }) {
    return (
        <div className="space-y-1">
            <span className="text-[10px] font-bold text-ink-soft uppercase tracking-wider">{label}</span>
            <div className={clsx("text-sm font-extrabold", color)}>{value}</div>
        </div>
    );
}

function HistoryItem({ date, time, title, desc, done }: any) {
    return (
        <div className="flex gap-4">
            <div className="flex flex-col items-center">
                <div className={clsx("w-3 h-3 rounded-full mt-1", done ? "bg-success" : "bg-line")} />
                <div className="w-[1px] flex-1 bg-line my-1" />
            </div>
            <div className="pb-6">
                <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs font-bold text-navy">{title}</span>
                    <span className="text-[10px] text-ink-soft">{date} - {time}</span>
                </div>
                <p className="text-[11px] text-ink-soft leading-relaxed">{desc}</p>
            </div>
        </div>
    );
}

function getStatusLabel(status: string) {
    const labels: any = {
        draft: "پیش‌نویس",
        map_in_progress: "نقشه‌برداری",
        claim_in_progress: "درج ادعا",
        cert_in_progress: "گواهی اقدام",
        completed: "تکمیل شده"
    };
    return labels[status] || status;
}

function getStatusColor(status: string) {
    if (status === 'completed') return "text-success";
    if (status === 'draft') return "text-ink-soft";
    return "text-brass";
}

function getStepGuide(status: string, mapStatus: string) {
    if (status === 'draft') return "در این مرحله باید اطلاعات هویتی و مشخصات دقیق ملک خود را تکمیل کنید تا پرونده جهت نقشه‌برداری به کارشناس ارجاع شود.";
    if (status === 'map_in_progress') {
        if (mapStatus === 'pending_expert_assignment') return "پرونده شما در صف تخصیص کارشناس ثبتی است. این فرآیند معمولاً کمتر از ۲۴ ساعت زمان می‌برد.";
        if (mapStatus === 'fieldwork_in_progress') return "کارشناس در حال انجام عملیات میدانی است. پس از ترسیم نقشه، باید آن را در این پنل تایید کنید.";
        return "سرویس نقشه در حال پردازش است.";
    }
    return "در حال بارگذاری راهنمای مرحله...";
}
