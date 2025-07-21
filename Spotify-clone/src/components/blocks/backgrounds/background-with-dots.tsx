import { cn } from "@/lib/utils";
import React from "react";

export function BackgroundDots() {
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
        "pointer-events-none absolute inset-0 z-0 h-full w-full",
        "bg-[radial-gradient(circle_at_0.5px_0.5px,rgba(0,0,0,0.2)_0.5px,transparent_0)]",
        "dark:bg-[radial-gradient(circle_at_0.5px_0.5px,rgba(255,255,255,0.2)_0.5px,transparent_0)]",
        "bg-repeat",
        "[background-size:8px_8px]"
      )}
    />
  );
};
