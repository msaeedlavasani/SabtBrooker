"use client";

import { useState } from "react";
import api from "@/lib/api";
import { Upload, CheckCircle2, Loader2, X } from "lucide-react";
import axios from "axios";

interface FileUploadProps {
  label: string;
  onUploadSuccess: (fileId: string) => void;
  accept?: string;
}

export default function FileUpload({ label, onUploadSuccess, accept = "image/*" }: FileUploadProps) {
  const [loading, setLoading] = useState(false);
  const [fileId, setFileId] = useState("");
  const [error, setError] = useState("");

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setLoading(true);
    setError("");

    try {
      // 1. Get Presigned URL
      const res = await api.get(`/v1/storage/presigned-url?name=${encodeURIComponent(file.name)}`);
      const { upload_url, file_id } = res.data;

      // 2. Upload directly to S3/MinIO
      await axios.put(upload_url, file, {
        headers: {
          'Content-Type': file.type,
        },
      });

      setFileId(file_id);
      onUploadSuccess(file_id);
    } catch (err: any) {
      setError("خطا در بارگذاری فایل");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const reset = () => {
    setFileId("");
    setError("");
  };

  return (
    <div className="space-y-2">
      <label className="text-[12px] font-bold text-ink-soft">{label}</label>
      
      {fileId ? (
        <div className="flex items-center justify-between rounded-lg border border-success bg-success-bg p-3 text-success">
          <div className="flex items-center gap-2 text-sm font-bold">
            <CheckCircle2 size={18} />
            فایل با موفقیت بارگذاری شد
          </div>
          <button onClick={reset} className="text-ink-soft hover:text-error">
            <X size={18} />
          </button>
        </div>
      ) : (
        <div className="relative">
          <input
            type="file"
            accept={accept}
            onChange={handleFileChange}
            disabled={loading}
            className="absolute inset-0 z-10 h-full w-full cursor-pointer opacity-0 disabled:cursor-not-allowed"
          />
          <div className={`flex items-center justify-center gap-3 rounded-lg border-2 border-dashed p-4 transition-colors ${
            loading ? "bg-gray-50 border-line" : "border-line hover:border-navy hover:bg-gray-50"
          }`}>
            {loading ? (
              <Loader2 className="animate-spin text-navy" size={24} />
            ) : (
              <>
                <Upload size={20} className="text-ink-soft" />
                <span className="text-sm font-bold text-ink-soft">انتخاب فایل و بارگذاری</span>
              </>
            )}
          </div>
        </div>
      )}
      {error && <p className="text-[11px] text-error font-bold">{error}</p>}
    </div>
  );
}
