import { cn } from "@/lib/utils";
import React from "react";

export function BackgroundWithSkewedLines() {
  return (
    <div className="flex h-screen flex-col items-center justify-center">
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
    <div
      className={cn(
        "pointer-events-none absolute inset-0 z-0 h-full w-full overflow-hidden [mask-image:linear-gradient(to_bottom,white,transparent,white)]"
      )}
    >
      <svg
        className="absolute h-[200%] w-[200%] -translate-x-1/2 -translate-y-1/2"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <pattern
            id="skewed-lines"
            width="40"
            height="40"
            patternUnits="userSpaceOnUse"
          >
            <path
              d="M-10,30 L30,-10 M-20,40 L40,-20 M-10,50 L50,-10"
              className="stroke-neutral-800/20 dark:stroke-neutral-200/20"
              strokeWidth="1"
              fill="none"
            />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#skewed-lines)" />
      </svg>
    </div>
  );
};
