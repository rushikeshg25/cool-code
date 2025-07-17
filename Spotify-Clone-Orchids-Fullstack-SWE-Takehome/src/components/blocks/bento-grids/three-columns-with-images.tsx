"use client";

import { cn } from "@/lib/utils";
import React from "react";
import { motion } from "motion/react";
import Image from "next/image";

export function ThreeColumnsWithImages() {
  return (
    <div className="mx-auto my-20 w-full max-w-7xl px-4 md:px-8">
      <h2 className="text-bold text-neutral-8000 text-center font-sans text-xl font-bold tracking-tight md:text-4xl dark:text-neutral-100">
        Dashboard for winners
      </h2>
      <p className="mx-auto mt-4 max-w-lg text-center text-sm text-neutral-600 dark:text-neutral-400">
        Our state of the art dashboard features the most comprehensive insights
        for your SaaS tool, with a blazing fast, we are so back AI text
        generation model.
      </p>
      <div className="cols-1 mt-20 grid gap-4 md:auto-rows-[25rem] md:grid-cols-3">
        {items.map((item, index) => (
          <Card
            key={item.title + index + "card-three-columns-with-images"}
            className={cn("flex flex-col justify-between", item.className)}
          >
            <CardContent className="h-40">
              <CardTitle>{item.title}</CardTitle>
              <CardDescription>{item.description}</CardDescription>
            </CardContent>
            <CardSkeletonBody>{item.header}</CardSkeletonBody>
          </Card>
        ))}
      </div>
    </div>
  );
}

const items = [
  {
    title: "Dashboard that matters",
    description:
      "Discover insights and trends with our advanced analytics dashboard.",
    header: (
      <Image
        src="https://assets.aceternity.com/pro/bento-1.png"
        alt="Bento grid 1"
        width={500}
        height={500}
        className="ml-6 w-full rounded-lg object-cover"
      />
    ),
    className: "md:col-span-1",
  },
  {
    title: "Automated emails",
    description:
      "Send emails in bulk to everyone, with AI-powered suggestions.",
    header: (
      <Image
        src="https://assets.aceternity.com/pro/bento-2.png"
        alt="Bento grid 2"
        width={500}
        height={500}
      />
    ),
    className: "md:col-span-1",
  },
  {
    title: "Super fast Analytics",
    description:
      "Get insights on your data with our blazing fast analytics dashboard.",
    header: (
      <Image
        src="https://assets.aceternity.com/pro/bento-4.png"
        alt="Bento grid 3"
        width={500}
        height={500}
        className="-ml-6 -mt-4 rounded-lg object-cover md:-mt-0"
      />
    ),
    className: "md:col-span-1",
  },
  {
    title: "Admin portal",
    description: "Manage your data with our admin portal.",
    header: (
      <div className="flex">
        <Image
          src="https://assets.aceternity.com/pro/bento-4.png"
          alt="Bento grid 4"
          width={500}
          height={500}
          className="ml-6 rounded-lg object-cover"
        />
        <Image
          src="https://assets.aceternity.com/pro/bento-4.png"
          alt="Bento grid 4"
          width={500}
          height={500}
          className="ml-6 mt-8 rounded-lg object-cover"
        />
      </div>
    ),
    className: "md:col-span-2",
  },

  {
    title: "99.99% uptime SLA",
    description: "We guarantee 99.99% uptime SLA for our platform.",
    header: (
      <Image
        src="https://assets.aceternity.com/pro/bento-5.png"
        alt="Bento grid 5"
        width={500}
        height={500}
        className="mx-auto -mb-4 w-[calc(100%-3rem)] rounded-lg"
      />
    ),
    className: "md:col-span-1",
  },
];

// Card structure
const CardSkeletonBody = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return <div className={cn("overflow-hidden", className)}>{children}</div>;
};

const CardContent = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return <div className={cn("p-6", className)}>{children}</div>;
};

const CardTitle = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  const variants = { initial: { x: 0 }, animate: { x: 10 } };
  return (
    <motion.h3
      variants={variants}
      transition={{ type: "easeOut", duration: 0.2 }}
      className={cn(
        "font-sans text-base font-medium tracking-tight text-neutral-700 dark:text-neutral-100",
        className
      )}
    >
      {children}
    </motion.h3>
  );
};
const CardDescription = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  const variants = { initial: { x: 0 }, animate: { x: 15 } };
  return (
    <motion.p
      variants={variants}
      transition={{ type: "easeOut", duration: 0.2 }}
      className={cn(
        "mt-2 max-w-xs font-sans text-base font-normal tracking-tight text-neutral-500 dark:text-neutral-400",
        className
      )}
    >
      {children}
    </motion.p>
  );
};

const Card = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return (
    <motion.div
      whileHover="animate"
      className={cn(
        "group isolate flex flex-col overflow-hidden rounded-2xl bg-white shadow-[0_1px_1px_rgba(0,0,0,0.05),0_4px_6px_rgba(34,42,53,0.04),0_24px_68px_rgba(47,48,55,0.05),0_2px_3px_rgba(0,0,0,0.04)] dark:bg-neutral-900",
        className
      )}
    >
      {children}
    </motion.div>
  );
};
