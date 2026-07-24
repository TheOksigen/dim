import type { Metadata } from "next";
import type { ReactNode } from "react";
import "./styles.css";

export const metadata: Metadata = {
  title: "İmtahan nəticələrinin yoxlanılması",
  description: "FIN kodu ilə imtahan nəticələrinin təhlükəsiz yoxlanılması",
};

export default function RootLayout({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="az">
      <body>{children}</body>
    </html>
  );
}
