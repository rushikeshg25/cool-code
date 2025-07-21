import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "McBeats",
  description: "Your music streaming companion",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className="antialiased bg-[#121212] text-white">{children}</body>
    </html>
  );
}