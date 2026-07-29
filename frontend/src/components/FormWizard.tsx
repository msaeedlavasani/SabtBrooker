
"use client";

import React, { useState, useEffect } from 'react';
import { Screen, Field } from '@/lib/workflow-configs';
import { CheckCircle2, Upload, Plus, X, Loader2, AlertTriangle, ShieldCheck } from 'lucide-react';
import { clsx } from 'clsx';
import api from '@/lib/api';

interface FormWizardProps {
  screens: Screen[];
  initialData?: any;
  onComplete: (data: any) => void;
  isLoading?: boolean;
}

export default function FormWizard({ screens, initialData = {}, onComplete, isLoading }: FormWizardProps) {
  const [data, setData] = useState(initialData);
  const [currentScreenIdx, setCurrentScreenIdx] = useState(0);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [otpSent, setOtpSent] = useState(false);
  const [otpValue, setOtpValue] = useState("");
  const [devOtp, setDevOtp] = useState("");

  const currentScreen = screens[currentScreenIdx];

  const handleFieldChange = (key: string, value: any) => {
    setData((prev: any) => ({ ...prev, [key]: value }));
    if (errors[key]) {
      setErrors((prev) => {
        const next = { ...prev };
        delete next[key];
        return next;
      });
    }
  };

  const handleRepeatAdd = (key: string, subFields: Field[]) => {
    const items = data[key] || [];
    const newItem = subFields.reduce((acc: any, f) => ({ ...acc, [f.key]: null }), {});
    setData((prev: any) => ({ ...prev, [key]: [...items, newItem] }));
  };

  const handleRepeatRemove = (key: string, idx: number) => {
    const items = [...(data[key] || [])];
    items.splice(idx, 1);
    setData((prev: any) => ({ ...prev, [key]: items }));
  };

  const handleRepeatFieldChange = (rootKey: string, idx: number, key: string, value: any) => {
    const items = [...(data[rootKey] || [])];
    items[idx] = { ...items[idx], [key]: value };
    setData((prev: any) => ({ ...prev, [rootKey]: items }));
  };

  const validateScreen = () => {
    const newErrors: Record<string, string> = {};
    if (!currentScreen.fields) return true;

    currentScreen.fields.forEach(f => {
      if (f.visibleIf && !f.visibleIf(data)) return;
      
      const val = data[f.key];
      if (f.required && (val === undefined || val === null || val === "" || val === false)) {
        newErrors[f.key] = "این فیلد الزامی است";
      }
      
      if (f.validate) {
        const customErr = f.validate(val, data);
        if (customErr) newErrors[f.key] = customErr;
      }
    });

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const next = async () => {
    if (validateScreen()) {
      if (currentScreen.kind === 'otp' && !otpSent) {
        // Send OTP logic (mock for now or call real API)
        try {
            const res = await api.post("/v1/auth/otp/send", { mobile: data.applicant_phone });
            if (res.data.dev_otp) setDevOtp(res.data.dev_otp);
            setOtpSent(true);
        } catch (e) {
            alert("خطا در ارسال رمز یکبار مصرف");
        }
        return;
      }

      if (currentScreenIdx < screens.length - 1) {
        setCurrentScreenIdx(currentScreenIdx + 1);
        setOtpSent(false);
        setOtpValue("");
      } else {
        onComplete(data);
      }
    }
  };

  const back = () => {
    if (currentScreenIdx > 0) {
      setCurrentScreenIdx(currentScreenIdx - 1);
      setOtpSent(false);
      setOtpValue("");
    }
  };

  if (currentScreen.kind === 'result') {
    return (
        <div className="space-y-6 text-center py-10">
            <div className="mx-auto w-24 h-24 rounded-full bg-success-bg flex items-center justify-center text-success animate-bounce">
                <CheckCircle2 size={48} />
            </div>
            <h2 className="text-2xl font-bold text-navy">{currentScreen.title}</h2>
            <div className="p-6 bg-paper rounded-xl border border-line inline-block">
                <p className="text-sm text-ink-soft mb-2">{currentScreen.label}</p>
                <p className="text-3xl font-mono font-bold text-navy tracking-wider">
                    {currentScreen.prefix}-{Math.floor(100000 + Math.random() * 899999)}
                </p>
            </div>
            <p className="text-sm text-ink-soft">این کد رهگیری را برای مراحل بعدی نزد خود نگه دارید.</p>
            <button onClick={() => onComplete(data)} className="btn-primary px-8">تایید و بازگشت به داشبورد</button>
        </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-xl font-bold text-navy">{currentScreen.title}</h2>
            {currentScreen.note && <p className="text-xs text-brass font-semibold mt-1">{currentScreen.note}</p>}
        </div>
        <div className="text-xs text-ink-soft font-bold bg-gray-100 px-3 py-1 rounded-full">
            گام {currentScreenIdx + 1} از {screens.length}
        </div>
      </div>

      {currentScreen.warn && (
        <div className="rounded-lg bg-warning-bg border border-warning-border p-4 text-xs text-warning-text leading-relaxed flex gap-3">
            <AlertTriangle size={16} className="shrink-0" />
            {currentScreen.warn}
        </div>
      )}

      {currentScreen.kind === 'otp' ? (
        <div className="space-y-6 py-4">
            {!otpSent ? (
                <div className="text-center space-y-4">
                    <p className="text-sm text-ink-soft">برای ثبت نهایی و تایید رضایت، رمز یکبار مصرف به شماره {data.applicant_phone} ارسال خواهد شد.</p>
                    {currentScreen.requireAck && (
                        <div className="flex items-center gap-2 justify-center">
                            <input 
                                type="checkbox" 
                                id="ack" 
                                checked={!!data[currentScreen.requireAck]} 
                                onChange={(e) => handleFieldChange(currentScreen.requireAck!, e.target.checked)}
                            />
                            <label htmlFor="ack" className="text-xs font-bold text-navy cursor-pointer">موارد فوق را مطالعه کرده و می‌پذیرم</label>
                        </div>
                    )}
                </div>
            ) : (
                <div className="max-w-xs mx-auto space-y-4">
                    <label className="text-sm font-bold text-ink block text-center">رمز ۵ رقمی ارسال شده را وارد کنید</label>
                    <input 
                        type="text" 
                        maxLength={5}
                        className="w-full p-4 text-center text-2xl font-mono tracking-[1rem] border border-line rounded-xl focus:border-navy focus:outline-none"
                        value={otpValue}
                        onChange={(e) => setOtpValue(e.target.value)}
                    />
                    {devOtp && <p className="text-[11px] text-brass text-center">محیط توسعه: کد تایید {devOtp} است</p>}
                </div>
            )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {currentScreen.fields?.map(f => {
                if (f.visibleIf && !f.visibleIf(data)) return null;
                return (
                    <div key={f.key} className={clsx("field space-y-1", (f.type === 'textarea' || f.type === 'repeat') && "md:col-span-2")}>
                        <div className="flex items-center gap-2">
                            <label className="text-sm font-semibold text-ink">{f.label}</label>
                            {f.ref && <span className="text-[10px] bg-brass/10 text-brass px-1.5 rounded">{f.ref}</span>}
                        </div>
                        
                        {renderField(f, data, handleFieldChange, handleRepeatFieldChange, handleRepeatAdd, handleRepeatRemove, errors[f.key])}
                        
                        {f.hint && <p className="text-[11px] text-ink-soft">{f.hint}</p>}
                        {f.chainNote && <p className="text-[11px] text-success font-semibold">{f.chainNote}</p>}
                        {errors[f.key] && <p className="text-[11px] text-error font-bold">{errors[f.key]}</p>}
                    </div>
                );
            })}
        </div>
      )}

      <div className="flex justify-between pt-6 border-t border-line">
        <button 
            onClick={back} 
            disabled={currentScreenIdx === 0 || isLoading}
            className="px-6 py-2 rounded-lg border border-line text-sm font-bold text-ink-soft disabled:opacity-30"
        >
            قبلی
        </button>
        <button 
            onClick={next} 
            disabled={isLoading || (currentScreen.kind === 'otp' && otpSent && otpValue.length < 5) || (currentScreen.kind === 'otp' && !otpSent && currentScreen.requireAck && !data[currentScreen.requireAck])}
            className="px-8 py-2 rounded-lg bg-navy text-white text-sm font-bold hover:opacity-90 disabled:opacity-50 flex items-center gap-2"
        >
            {isLoading ? <Loader2 size={16} className="animate-spin" /> : (currentScreenIdx === screens.length - 1 ? 'ثبت نهایی' : 'بعدی')}
        </button>
      </div>
    </div>
  );
}

function renderField(
    f: Field, 
    data: any, 
    onChange: (k: string, v: any) => void,
    onRepeatChange: (rk: string, idx: number, k: string, v: any) => void,
    onRepeatAdd: (k: string, sf: Field[]) => void,
    onRepeatRemove: (k: string, idx: number) => void,
    error?: string
) {
    const val = data[f.key];

    switch (f.type) {
        case 'text':
        case 'number':
        case 'date':
            return (
                <input 
                    type={f.type === 'text' ? 'text' : (f.type === 'number' ? 'number' : 'date')}
                    value={val || ''}
                    onChange={(e) => onChange(f.key, f.type === 'number' ? Number(e.target.value) : e.target.value)}
                    className={clsx(
                        "w-full p-2.5 text-sm border rounded-lg focus:border-navy focus:outline-none bg-gray-50/50 font-mono",
                        error ? "border-error" : "border-line"
                    )}
                    readOnly={f.readonly}
                />
            );
        case 'select':
            return (
                <select 
                    value={val || ''}
                    onChange={(e) => onChange(f.key, e.target.value)}
                    className={clsx(
                        "w-full p-2.5 text-sm border rounded-lg focus:border-navy focus:outline-none bg-gray-50/50",
                        error ? "border-error" : "border-line"
                    )}
                >
                    <option value="">— انتخاب کنید —</option>
                    {f.options?.map(o => <option key={o.v} value={o.v}>{o.l}</option>)}
                </select>
            );
        case 'checkbox':
            return (
                <div className="flex items-center gap-2 py-1">
                    <input 
                        type="checkbox" 
                        id={f.key}
                        checked={!!val}
                        onChange={(e) => onChange(f.key, e.target.checked)}
                        className="w-4 h-4"
                    />
                    <label htmlFor={f.key} className="text-xs font-semibold text-ink-soft cursor-pointer">تایید می‌کنم</label>
                </div>
            );
        case 'textarea':
            return (
                <textarea 
                    value={val || ''}
                    onChange={(e) => onChange(f.key, e.target.value)}
                    rows={3}
                    className={clsx(
                        "w-full p-2.5 text-sm border rounded-lg focus:border-navy focus:outline-none bg-gray-50/50",
                        error ? "border-error" : "border-line"
                    )}
                />
            );
        case 'file':
            return (
                <div 
                    onClick={() => onChange(f.key, !val)}
                    className={clsx(
                        "flex items-center justify-center gap-2 p-3 border-2 border-dashed rounded-xl cursor-pointer transition-colors",
                        val ? "bg-success-bg border-success text-success" : "bg-gray-50 border-line text-ink-soft hover:bg-gray-100"
                    )}
                >
                    {val ? <CheckCircle2 size={18} /> : <Upload size={18} />}
                    <span className="text-xs font-bold">{val ? 'بارگذاری شد' : 'انتخاب فایل'}</span>
                </div>
            );
        case 'repeat':
            const items = data[f.key] || [];
            return (
                <div className="space-y-4">
                    <div className="space-y-3">
                        {items.map((item: any, idx: number) => (
                            <div key={idx} className="relative p-4 border border-line rounded-xl bg-gray-50/30">
                                <button 
                                    onClick={() => onRepeatRemove(f.key, idx)}
                                    className="absolute left-2 top-2 text-error hover:bg-error-bg p-1 rounded"
                                >
                                    <X size={14} />
                                </button>
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
                                    {f.subFields?.map(sf => (
                                        <div key={sf.key} className="space-y-1">
                                            <label className="text-xs font-bold text-ink-soft">{sf.label}</label>
                                            {renderField(
                                                sf, 
                                                { [sf.key]: item[sf.key] }, 
                                                (k, v) => onRepeatChange(f.key, idx, k, v),
                                                () => {}, () => {}, () => {}
                                            )}
                                        </div>
                                    ))}
                                </div>
                            </div>
                        ))}
                    </div>
                    <button 
                        onClick={() => onRepeatAdd(f.key, f.subFields || [])}
                        className="flex items-center gap-2 text-xs font-bold text-navy bg-navy/5 px-4 py-2 rounded-lg hover:bg-navy/10"
                    >
                        <Plus size={14} />
                        {f.addLabel || 'افزودن مورد جدید'}
                    </button>
                </div>
            );
        default:
            return null;
    }
}
