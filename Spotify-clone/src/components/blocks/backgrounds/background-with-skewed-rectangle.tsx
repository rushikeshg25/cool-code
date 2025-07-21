import { cn } from "@/lib/utils";
import React from "react";

export function BackgroundWithSkewedRectangles() {
  return (
    <div className="flex h-screen flex-col items-center justify-center border-b border-neutral-100 dark:border-neutral-800">
      <Background />
      <Content />
    </div>
  );
}

const Content = () => {
  return (
    <div className="relative z-10">
      <h1 className="text-balance mx-auto max-w-2xl text-center text-3xl font-bold text-black dark:text-white md:text-5xl">
        Web apps that make you feel like you&apos;re in the future
      </h1>
      <p className="text-balance mx-auto mt-4 max-w-2xl text-center text-base text-neutral-800 dark:text-neutral-200">
        We are a team of developers who are passionate about creating web apps
        that make you feel like you&apos;re in the future.
      </p>
      <div className="mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
        <button className="w-40 rounded-lg bg-gradient-to-b from-neutral-950 to-neutral-800 px-4 py-2 text-sm text-white">
          Sign up
        </button>
        <button className="w-40 rounded-lg bg-gradient-to-b from-neutral-100 to-neutral-50 px-4 py-2 text-sm text-black">
          Register now
        </button>
      </div>
    </div>
  );
};

const Background = () => {
  return (
    <div className="pointer-events-none absolute inset-0 z-0 h-full w-full overflow-hidden [perspective:1000px] [transform-style:preserve-3d]">
      <Rectangles
        style={{ transform: "rotateX(45deg)" }}
        className="[mask-image:linear-gradient(to_top,white,transparent)]"
      />
      <Rectangles
        style={{ transform: "rotateX(-45deg)" }}
        className="[mask-image:linear-gradient(to_bottom,white,transparent)]"
      />
    </div>
  );
};

const Rectangles = ({
  className,
  ...props
}: {
  className?: string;
  style?: React.CSSProperties;
}) => {
  const rectangleSVGLight = `<svg width='40' height='40' viewBox='0 0 40 40' xmlns='http://www.w3.org/2000/svg'><rect width='40' height='40' x='0' y='0' stroke='rgba(0,0,0,0.1)' fill='none' /></svg>`;
  const rectangleSVGDark = `<svg width='40' height='40' viewBox='0 0 40 40' xmlns='http://www.w3.org/2000/svg'><rect width='40' height='40' x='0' y='0' stroke='rgba(255,255,255,0.15)' fill='none' /></svg>`;
  const encodedRectangleSVGLight = encodeURIComponent(rectangleSVGLight);
  const encodedRectangleSVGDark = encodeURIComponent(rectangleSVGDark);
  return (
    <div
      className={cn(
        "pointer-events-none absolute inset-0 z-0 h-full w-full overflow-hidden",
        className
      )}
      {...props}
    >
      <div
        className={cn("h-full w-full dark:hidden")}
        style={{
          backgroundImage: `url("data:image/svg+xml,${encodedRectangleSVGLight}")`,
          backgroundSize: "40px 40px",
          backgroundPosition: "center",
          backgroundRepeat: "repeat",
        }}
      />
      <div
        className={cn("hidden h-full w-full dark:block")}
        style={{
          backgroundImage: `url("data:image/svg+xml,${encodedRectangleSVGDark}")`,
          backgroundSize: "40px 40px",
          backgroundPosition: "center",
          backgroundRepeat: "repeat",
        }}
      />
    </div>
  );
};
