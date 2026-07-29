import type { Metadata } from "next";
import { Vazirmatn } from "next/font/google";
import "./globals.css";

const vazir = Vazirmatn({
  variable: "--font-vazirmatn",
  subsets: ["arabic"],
});

export const metadata: Metadata = {
  title: "سامانه کارگزاری ماده ۱۰",
  description: "سامانه خدمات زنجیره‌ای سازمان ثبت اسناد و املاک",
  manifest: "/manifest.json",
  themeColor: "#101B33",
  appleWebApp: {
    capable: true,
    statusBarStyle: "default",
    title: "SabtBrooker",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="fa" dir="rtl" className={`${vazir.variable} h-full antialiased`}>
      <body className="min-h-full flex flex-col font-sans">{children}</body>
    </html>
  );
}
