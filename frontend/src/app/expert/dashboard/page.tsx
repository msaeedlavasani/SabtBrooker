"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import api from "@/lib/api";
import { 
  ClipboardCheck, 
  Map, 
  User, 
  LogOut,
  AlertCircle,
  CheckCircle2,
  Clock,
  ChevronLeft,
  FileSearch
} from "lucide-react";

export default function ExpertDashboard() {
  const [user, setUser] = useState<any>(null);
  const [pendingCases, setPendingCases] = useState([]);
  const [activeTasks, setActiveTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  const fetchData = async () => {
    try {
      const userRes = await api.get("/v1/auth/me");
      setUser(userRes.data);
      
      // For now, experts list all cases and we filter locally or via endpoint if available
      // In a real scenario, we'd have /v1/expert/cases/pending
      const casesRes = await api.get("/v1/cases");
      const allCases = casesRes.data || [];
      
      setPendingCases(allCases.filter((c: any) => c.status === "map_in_progress" && !c.survey_expert_id));
      setActiveTasks(allCases.filter((c: any) => c.survey_expert_id === userRes.data.id));
      
    } catch (err) {
      localStorage.removeItem("access_token");
      router.push("/");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [router]);

  const handleLogout = () => {
    localStorage.removeItem("access_token");
    router.push("/");
  };

  if (loading) return <div className="flex min-h-screen items-center justify-center bg-paper font-bold text-navy">در حال بارگذاری پنل کارشناسی...</div>;

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <header className="bg-navy p-5 text-white shadow-lg border-b-4 border-brass">
        <div className="mx-auto flex max-w-6xl items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="bg-brass p-2 rounded-lg">
              <ClipboardCheck size={24} className="text-navy" />
            </div>
            <div>
              <h1 className="text-xl font-bold">پنل کارشناس ثبت</h1>
              <p className="text-[11px] text-brass-light">
                کارشناس: {user?.first_name} {user?.last_name} ({user?.mobile})
              </p>
            </div>
          </div>
          <button 
            onClick={handleLogout}
            className="flex items-center gap-2 rounded-lg bg-white/10 px-4 py-2 text-sm hover:bg-white/20 transition-colors"
          >
            <LogOut size={18} />
            خروج
          </button>
        </div>
      </header>

      <main className="mx-auto max-w-6xl p-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          
          {/* Left Column: New Requests */}
          <div className="lg:col-span-2 space-y-6">
            <section className="bg-white rounded-2xl border border-line shadow-sm overflow-hidden">
              <div className="bg-gray-50 border-b border-line p-6 flex justify-between items-center">
                <h2 className="text-lg font-bold text-navy flex items-center gap-2">
                  <FileSearch size={20} className="text-brass" />
                  پرونده‌های جدید (منتظر کارشناس)
                </h2>
                <span className="bg-navy text-white text-xs px-3 py-1 rounded-full font-bold">
                  {pendingCases.length} مورد
                </span>
              </div>

              {pendingCases.length === 0 ? (
                <div className="p-20 text-center text-ink-soft">
                  <p>در حال حاضر پرونده جدیدی برای تخصیص وجود ندارد.</p>
                </div>
              ) : (
                <div className="divide-y divide-line">
                  {pendingCases.map((c: any) => (
                    <div key={c.id} className="p-6 hover:bg-gray-50 transition-colors flex justify-between items-center">
                      <div className="space-y-1">
                        <div className="font-bold text-navy">{c.province}، {c.city}</div>
                        <div className="text-xs text-ink-soft">{c.address_detail}</div>
                        <div className="text-[10px] font-mono opacity-60">ID: {c.id}</div>
                      </div>
                      <button 
                        onClick={async () => {
                          try {
                            // Find the map service ID for this case
                            // In a real app, we'd have the map service ID directly
                            // Here we assume case ID is what we need or we fetch it
                            await api.post(`/v1/map-services/${c.id}/accept`);
                            alert("پرونده با موفقیت پذیرفته شد");
                            fetchData();
                          } catch (err: any) {
                            alert(err.response?.data?.error || "خطا در پذیرش پرونده");
                          }
                        }}
                        className="bg-navy text-white px-6 py-2 rounded-lg text-sm font-bold hover:bg-navy-light transition-transform active:scale-95"
                      >
                        قبول پرونده
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </section>
          </div>

          {/* Right Column: My Active Tasks */}
          <div className="space-y-6">
            <section className="bg-white rounded-2xl border border-line shadow-sm overflow-hidden border-r-4 border-r-success">
              <div className="p-6 border-b border-line">
                <h2 className="text-lg font-bold text-navy flex items-center gap-2">
                  <Clock size={20} className="text-success" />
                  وظایف فعال من
                </h2>
              </div>
              
              {activeTasks.length === 0 ? (
                <div className="p-10 text-center text-ink-soft text-sm">
                  شما در حال حاضر وظیفه فعالی ندارید.
                </div>
              ) : (
                <div className="divide-y divide-line">
                  {activeTasks.map((c: any) => (
                    <div key={c.id} className="p-4 hover:bg-gray-50 cursor-pointer group">
                      <div className="flex justify-between items-start mb-2">
                        <span className="text-[10px] font-bold bg-success/10 text-success px-2 py-0.5 rounded">در جریان</span>
                        <ChevronLeft size={16} className="text-ink-soft group-hover:translate-x-[-4px] transition-transform" />
                      </div>
                      <div className="text-sm font-bold text-navy">{c.province} - {c.city}</div>
                    </div>
                  ))}
                </div>
              )}
            </section>

            <div className="bg-brass/10 border border-brass/20 rounded-2xl p-6 text-brass-ink">
              <h3 className="font-bold mb-2 flex items-center gap-2">
                <AlertCircle size={18} />
                نکته مهم کارشناسی
              </h3>
              <p className="text-xs leading-relaxed opacity-80">
                طبق ماده ۱۲ دستورالعمل، کارشناس موظف است حداکثر ظرف ۴۸ ساعت پس از قبول پرونده، نسبت به هماهنگی با متقاضی جهت بازدید میدانی اقدام نماید.
              </p>
            </div>
          </div>

        </div>
      </main>
    </div>
  );
}
