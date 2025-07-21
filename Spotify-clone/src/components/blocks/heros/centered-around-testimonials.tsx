"use client";
import { cn } from "@/lib/utils";
import { useMotionValueEvent, useScroll } from "motion/react";
import { Outfit } from "next/font/google";
import Image from "next/image";
import React, { useRef, useState } from "react";
import { motion } from "motion/react";

const outfit = Outfit({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
});

export function CenteredAroundTestimonials() {
  const containerRef = useRef<HTMLDivElement | null>(null);

  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end start"],
  });

  const [movedOut, setMovedOut] = useState(false);

  useMotionValueEvent(scrollYProgress, "change", (latest) => {
    setMovedOut(parseFloat(latest.toFixed(2)) > 0.04);
  });

  const TestimonialCard = ({
    testimonial,
    direction,
    className,
  }: {
    testimonial: Testimonial;
    direction?: "left" | "right";
    className?: string;
  }) => {
    return (
      <motion.div
        layoutId={`testimonial-${testimonial.name}`}
        key={String(movedOut)}
        className={cn(
          "absolute z-20 flex items-center gap-2 rounded-md bg-white p-4 opacity-20 shadow-lg lg:opacity-100 dark:bg-neutral-800",
          className
        )}
        animate={{
          y: movedOut ? 300 : 0,
          x: movedOut
            ? direction === "left"
              ? -400
              : direction === "right"
              ? 400
              : 0
            : 0,
          rotate: testimonial.rotate ?? 20,
        }}
        transition={{ type: "spring", stiffness: 50, damping: 20 }}
      >
        <Image
          src={testimonial.src}
          alt={testimonial.name}
          width={50}
          height={50}
          className="rounded-full"
        />
        <div>
          <h3 className="text-xs text-neutral-800 md:text-base dark:text-neutral-200">
            {testimonial.name}
          </h3>
          <p className="max-w-md text-[10px] text-neutral-600 md:text-sm dark:text-neutral-400">
            {testimonial.quote}
          </p>
        </div>
      </motion.div>
    );
  };
  return (
    <div ref={containerRef} className="relative w-full overflow-hidden">
      <TestimonialCard
        testimonial={testimonials[0]}
        className="-left-10 top-20"
        direction="left"
      />
      <TestimonialCard
        testimonial={testimonials[1]}
        className="-left-10 top-1/2 -translate-y-1/2"
        direction="left"
      />
      <TestimonialCard
        testimonial={testimonials[2]}
        className="-right-10 top-20"
        direction="right"
      />
      <TestimonialCard
        testimonial={testimonials[3]}
        className="-left-10 bottom-20"
        direction="left"
      />
      <TestimonialCard
        testimonial={testimonials[4]}
        className="-right-10 bottom-1/2 -translate-y-1/2"
        direction="right"
      />
      <TestimonialCard
        testimonial={testimonials[5]}
        className="-right-10 bottom-20"
        direction="right"
      />
      <div className="relative flex h-screen flex-col items-center justify-center overflow-hidden bg-gray-50 px-4 md:px-8 dark:bg-neutral-900">
        <div className="absolute inset-0 z-30 h-full w-full bg-white opacity-80 md:opacity-0 dark:bg-neutral-900" />
        <Image
          src="https://assets.aceternity.com/pro/logos/aceternity-ui.png"
          alt="Aceternity Logo"
          width={100}
          height={100}
          className="relative z-50 aspect-square w-40 object-contain filter dark:invert"
        />
        <h1
          className={cn(
            "relative z-50 mx-auto mt-10 max-w-5xl text-center text-lg font-semibold text-neutral-700 sm:text-2xl md:text-4xl lg:text-7xl dark:text-neutral-100",
            outfit.className
          )}
        >
          Join the biggest Image Generation Hackathon ever
        </h1>

        <form className="relative z-50 mx-auto mt-10 flex w-full max-w-xl rounded-full border border-neutral-100 bg-white px-2 py-1.5 shadow-[0px_1px_3px_0px_rgba(0,0,0,0.15)_inset] dark:border-neutral-800 dark:bg-neutral-950">
          <input
            type="text"
            placeholder="Enter your email"
            className="w-full border-none bg-transparent text-sm text-black ring-0 focus:outline-none focus:ring-0 dark:text-white"
          />
          <button className="rounded-full bg-black px-4 py-1 text-sm text-white shadow-[0px_-1px_0px_0px_#FFFFFF40_inset,_0px_1px_0px_0px_#FFFFFF40_inset] dark:bg-white dark:text-black">
            Submit
          </button>
        </form>
      </div>
    </div>
  );
}

interface Testimonial {
  src: string;
  quote: string;
  name: string;
  designation?: string;
  rotate?: number;
}

export const testimonials: Testimonial[] = [
  {
    name: "Manu Arora",
    quote: "Fantastic AI, highly recommend it.",
    src: "https://i.pravatar.cc/150?img=1",
    designation: "Tech Innovator & Entrepreneur",
    rotate: -20,
  },
  {
    name: "Tyler Durden",
    quote: "AI revolutionized my business model.",
    src: "https://i.pravatar.cc/150?img=2",
    designation: "Creative Director & Business Owner",
    rotate: -10,
  },
  {
    name: "Alice Johnson",
    quote: "Transformed the way I work!",
    src: "https://i.pravatar.cc/150?img=3",
    designation: "Senior Software Engineer",
    rotate: 20,
  },
  {
    name: "Bob Smith",
    quote: "Absolutely revolutionary, a game-changer.",
    src: "https://i.pravatar.cc/150?img=4",
    designation: "Industry Analyst",
    rotate: -10,
  },
  {
    name: "Cathy Lee",
    quote: "Improved my work efficiency and daily life.",
    src: "https://i.pravatar.cc/150?img=5",
    designation: "Product Manager",
    rotate: 10,
  },
  {
    name: "David Wright",
    quote: "It's like having a superpower!",
    src: "https://i.pravatar.cc/150?img=6",
    designation: "Research Scientist",
    rotate: 20,
  },
];
