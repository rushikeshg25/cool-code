"use client";
import { HiArrowRight } from "react-icons/hi2";
import { motion } from "motion/react";

import Image from "next/image";
import { cn } from "@/lib/utils";

export function HeroWithCenteredImage() {
  return (
    <div className="relative mx-auto flex min-h-screen max-w-7xl flex-col overflow-hidden pt-20 md:pt-40">
      <motion.div
        initial={{ y: 40, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ ease: "easeOut", duration: 0.5 }}
        className="flex justify-center"
      >
        <Badge
          onClick={() => {
            console.log("clicked");
          }}
        >
          We&apos;ve raised $69M seed funding
        </Badge>
      </motion.div>
      <motion.h1
        initial={{ y: 40, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ ease: "easeOut", duration: 0.5 }}
        className="relative z-10 mx-auto mt-6 max-w-6xl text-center text-2xl font-semibold text-black md:text-4xl lg:text-8xl dark:text-white"
      >
        One-stop solution for all your content needs
      </motion.h1>
      <motion.p
        initial={{ y: 40, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ ease: "easeOut", duration: 0.5, delay: 0.2 }}
        className="relative z-10 mx-auto mt-6 max-w-3xl text-center text-base text-neutral-700 md:text-xl dark:text-neutral-400"
      >
        Space Rocket Tech is a team of passionate and experienced professionals
        who are dedicated to helping you achieve your content marketing goals.
      </motion.p>
      <motion.div
        initial={{ y: 80, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ ease: "easeOut", duration: 0.5, delay: 0.4 }}
        className="relative z-10 mt-6 flex items-center justify-center gap-4"
      >
        <Button>Get started</Button>
        <Button
          variant="simple"
          href="#"
          className="group flex items-center space-x-2"
        >
          <span>Contact us</span>
          <HiArrowRight className="h-3 w-3 stroke-[1px] text-neutral-700 transition-transform duration-200 group-hover:translate-x-1 dark:text-neutral-300" />
        </Button>
      </motion.div>
      <div className="relative mt-20 rounded-[32px] border border-neutral-200 bg-neutral-100 p-4 dark:border-neutral-700 dark:bg-neutral-800">
        <div className="pointer-events-none absolute inset-x-0 bottom-0 h-40 w-full scale-[1.1] bg-gradient-to-b from-transparent via-white to-white dark:via-black/50 dark:to-black" />
        <div className="rounded-[24px] border border-neutral-200 bg-white p-2 dark:border-neutral-700 dark:bg-black">
          <Image
            src="https://assets.aceternity.com/pro/header.webp"
            alt="header"
            width={1920}
            height={1080}
            className="rounded-[20px]"
          />
        </div>
      </div>
    </div>
  );
}

export const Badge: React.FC<
  { children: React.ReactNode } & React.ComponentPropsWithoutRef<"button">
> = ({ children, ...props }) => {
  return (
    <button
      {...props}
      className="group relative mx-auto inline-block w-fit cursor-pointer rounded-full bg-neutral-50 p-px text-[10px] font-semibold leading-6 text-neutral-700 no-underline shadow-zinc-900 sm:text-xs md:shadow-2xl dark:bg-neutral-700 dark:text-neutral-300"
    >
      <span className="absolute inset-0 overflow-hidden rounded-full">
        <span className="absolute inset-0 rounded-full opacity-0 transition-opacity duration-500 group-hover:opacity-100" />
      </span>
      <div className="relative z-10 flex items-center space-x-2 rounded-full bg-neutral-100 px-4 py-1.5 ring-1 ring-white/10 dark:bg-neutral-800">
        <span>{children}</span>
        <svg
          fill="none"
          height="16"
          viewBox="0 0 24 24"
          width="16"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M10.75 8.75L14.25 12L10.75 15.25"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="1.5"
          />
        </svg>
      </div>
      <span className="absolute -bottom-0 left-[1.125rem] h-px w-[calc(100%-2.25rem)] bg-gradient-to-r from-neutral-400/0 via-neutral-400/90 to-neutral-400/0 transition-opacity duration-500 group-hover:opacity-40" />
    </button>
  );
};

export const Button: React.FC<{
  children?: React.ReactNode;
  className?: string;
  variant?: "simple" | "outline" | "primary";
  as?: React.ElementType<any>;
  [x: string]: any;
}> = ({
  children,
  className,
  variant = "primary",
  as: Tag = "button" as any,
  ...props
}) => {
  const variantClass =
    variant === "simple"
      ? "bg-black relative z-10 bg-transparent hover:bg-gray-100  border border-transparent text-black text-sm md:text-sm transition font-medium duration-200  rounded-full px-4 py-2  flex items-center justify-center dark:text-white dark:hover:bg-neutral-800 dark:hover:shadow-xl"
      : variant === "outline"
      ? "bg-white relative z-10 hover:bg-black/90 hover:shadow-xl  text-black border border-black hover:text-white text-sm md:text-sm transition font-medium duration-200  rounded-full px-4 py-2  flex items-center justify-center"
      : variant === "primary"
      ? "bg-neutral-900 relative z-10 hover:bg-black/90  border border-transparent text-white text-sm md:text-sm transition font-medium duration-200  rounded-full px-4 py-2  flex items-center justify-center shadow-[0px_-1px_0px_0px_#FFFFFF40_inset,_0px_1px_0px_0px_#FFFFFF40_inset]"
      : "";
  return (
    <Tag
      className={cn(
        "relative z-10 flex items-center justify-center rounded-full bg-black px-4 py-2 text-sm font-medium text-white transition duration-200 hover:bg-black/90 md:text-sm",
        variantClass,
        className
      )}
      {...props}
    >
      {children ?? `Get Started`}
    </Tag>
  );
};
