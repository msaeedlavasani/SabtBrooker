"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import api from "@/lib/api";
import { 
  LayoutDashboard, 
  FileText, 
  Map, 
  ClipboardCheck, 
  LogOut,
  PlusCircle,
  Clock,
  CheckCircle2,
  AlertCircle
} from "lucide-react";

export default function Dashboard() {
  const [user, setUser] = useState<any>(null);
  const [cases, setCases] = useState([]);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const fetchData = async () => {
      try {
        const userRes = await api.get("/v1/auth/me");
        setUser(userRes.data);
        const casesRes = await api.get("/v1/cases");
        setCases(casesRes.data || []);
      } catch (err) {
        localStorage.removeItem("access_token");
        router.push("/");
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [router]);

  const handleLogout = () => {
    localStorage.removeItem("access_token");
    router.push("/");
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-paper text-navy">
        <div className="text-lg font-bold">در حال بارگذاری...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-paper pb-20">
      {/* Topbar */}
      <header className="bg-navy p-5 text-white shadow-lg">
        <div className="mx-auto flex max-w-6xl items-center justify-between">
          <div>
            <h1 className="text-xl font-bold">سامانه کارگزاری ماده ۱۰</h1>
            <p className="text-[11px] text-brass-light opacity-80">
              خوش آمدید، {user?.mobile}
            </p>
          </div>
          <button 
            onClick={handleLogout}
            className="flex items-center gap-2 rounded-lg bg-white/10 px-4 py-2 text-sm hover:bg-white/20"
          >
            <LogOut size={18} />
            خروج
          </button>
        </div>
      </header>

      <main className="mx-auto max-w-6xl p-6">
        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {/* Sidebar / Stats */}
          <div className="space-y-6">
            <div className="rounded-xl border border-line bg-white p-6 shadow-sm">
              <h2 className="mb-4 flex items-center gap-2 text-lg font-bold text-navy">
                <LayoutDashboard size={20} />
                وضعیت کلی
              </h2>
              <div className="space-y-4">
                <StatItem icon={<FileText size={16} />} label="کل پرونده‌ها" value={cases.length} />
                <StatItem icon={<Clock size={16} />} label="در انتظار بررسی" value={0} color="text-brass" />
                <StatItem icon={<CheckCircle2 size={16} />} label="تایید شده" value={0} color="text-success" />
              </div>
            </div>
            
            <button className="flex w-full items-center justify-center gap-2 rounded-xl bg-navy p-4 font-bold text-white shadow-md transition-transform hover:scale-[1.02]">
              <PlusCircle size={20} />
              ایجاد پرونده جدید
            </button>
          </div>

          {/* Main Content / Case List */}
          <div className="md:col-span-2">
            <div className="rounded-xl border border-line bg-white shadow-sm overflow-hidden">
              <div className="border-b border-line bg-gray-50 p-6">
                <h2 className="text-lg font-bold text-navy">آخرین پرونده‌ها</h2>
              </div>
              
              {cases.length === 0 ? (
                <div className="flex flex-col items-center justify-center p-20 text-ink-soft">
                  <AlertCircle size={48} className="mb-4 opacity-20" />
                  <p>هنوز پرونده‌ای ثبت نکرده‌اید.</p>
                </div>
              ) : (
                <div className="divide-y divide-line">
                  {cases.map((c: any) => (
                    <CaseRow key={c.id} c={c} />
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

function StatItem({ icon, label, value, color = "text-navy" }: any) {
  return (
    <div className="flex items-center justify-between border-b border-gray-50 pb-2">
      <div className="flex items-center gap-2 text-sm text-ink-soft">
        {icon}
        {label}
      </div>
      <div className={`font-mono font-bold ${color}`}>{value}</div>
    </div>
  );
}

function CaseRow({ c }: any) {
  const statusLabels: any = {
    draft: { label: "پیش‌نویس", color: "bg-gray-100 text-gray-700" },
    map_in_progress: { label: "نقشه‌برداری", color: "bg-brass/10 text-brass" },
    claim_in_progress: { label: "درج ادعا", color: "bg-navy/10 text-navy" },
    cert_in_progress: { label: "گواهی اقدام", color: "bg-navy/10 text-navy" },
    completed: { label: "تکمیل شده", color: "bg-success-bg text-success" },
  };

  const status = statusLabels[c.status] || { label: c.status, color: "bg-gray-100" };

  return (
    <div className="flex items-center justify-between p-6 hover:bg-gray-50 transition-colors">
      <div className="space-y-1">
        <div className="flex items-center gap-3">
          <span className={`rounded-full px-3 py-1 text-[10px] font-bold ${status.color}`}>
            {status.label}
          </span>
          <span className="font-mono text-[11px] text-ink-soft">
            {c.id.substring(0, 8)}...
          </span>
        </div>
        <div className="text-sm font-semibold text-navy">
          {c.province}، {c.city}
        </div>
        <div className="text-[11px] text-ink-soft">
          آخرین به‌روزرسانی: {new Date(c.updated_at || c.created_at).toLocaleDateString('fa-IR')}
        </div>
      </div>
      <button className="rounded-lg border border-line px-4 py-2 text-xs font-bold text-navy hover:bg-white hover:shadow-sm">
        مشاهده جزئیات
      </button>
    </div>
  );
}
