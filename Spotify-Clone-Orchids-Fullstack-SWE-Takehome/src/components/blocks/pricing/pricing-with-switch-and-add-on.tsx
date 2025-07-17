"use client";
import React from "react";
import {
  IconCircleCheckFilled,
  IconMessageCircleQuestion,
} from "@tabler/icons-react";
import { cn } from "@/lib/utils";
import { useState } from "react";
import { motion } from "motion/react";

export function PricingWithSwitchAndAddOn() {
  return (
    <div className="relative mx-auto flex max-w-7xl flex-col items-center justify-between border border-neutral-200 py-10 md:py-20 dark:border-neutral-800">
      <Icon className="absolute -left-4 -top-4" />
      <Icon className="absolute -right-4 -top-4" />
      <Icon className="absolute -bottom-4 -left-4" />
      <Icon className="absolute -bottom-4 -right-4" />
      <div className="relative w-full px-4 md:px-8">
        <h2 className="bg-gradient-to-b from-neutral-900 to-neutral-700 bg-clip-text text-center text-xl font-bold tracking-tight text-transparent md:text-4xl dark:from-white dark:to-neutral-200">
          Choose the plan that suits your needs
        </h2>
        <p className="mx-auto mt-4 max-w-2xl text-center text-sm font-normal text-neutral-700 md:text-base dark:text-neutral-400">
          Pick a plan that suits your needs and get started instantly.
        </p>
      </div>
      <Pricing />
    </div>
  );
}

export function Pricing() {
  const [active, setActive] = useState("monthly");
  const tabs = [
    { name: "Monthly", value: "monthly" },
    { name: "Yearly", value: "yearly" },
  ];

  return (
    <div className="relative">
      <div className="border-b">
        <div className="mx-auto mb-12 mt-10 flex w-fit items-center justify-center overflow-hidden rounded-full bg-neutral-100 dark:bg-neutral-800">
          {tabs.map((tab) => (
            <button
              key={tab.value}
              className={cn(
                "relative rounded-md px-4 py-2 text-sm font-medium text-gray-500 dark:text-neutral-400",
                active === tab.value ? "text-white dark:text-white" : ""
              )}
              onClick={() => setActive(tab.value)}
            >
              {active === tab.value && (
                <motion.span
                  layoutId="moving-div"
                  transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                  className="absolute inset-0 bg-indigo-700 dark:bg-indigo-700"
                />
              )}
              <span className="relative z-10">{tab.name}</span>
            </button>
          ))}
        </div>
      </div>
      <div className="relative z-20 mx-auto mt-4 grid grid-cols-1 items-center border-b border-t border-neutral-200 bg-neutral-100 px-4 md:mt-20 md:grid-cols-2 md:px-8 xl:grid-cols-3 dark:border-neutral-800 dark:bg-neutral-950">
        {tiers.map((tier, tierIdx) => (
          <div
            key={tier.id}
            className={cn(
              "flex h-full flex-col justify-between bg-neutral-100 px-6 py-8 sm:mx-8 lg:mx-0 dark:bg-neutral-950",
              tier.featured &&
                "relative bg-white shadow-2xl dark:bg-neutral-900"
            )}
          >
            <div className="">
              <h3
                id={tier.id}
                className={cn(
                  "text-base font-semibold leading-7 text-neutral-700 dark:text-neutral-200",
                  tier.featured && "text-black dark:text-white"
                )}
              >
                {tier.name}
              </h3>
              <p className="mt-4">
                <motion.span
                  initial={{ x: -20, opacity: 0 }}
                  animate={{ x: 0, opacity: 1 }}
                  transition={{
                    duration: 0.2,
                    ease: "easeOut",
                    delay: 0.1 * tierIdx,
                  }}
                  key={active}
                  className={cn(
                    "inline-block text-4xl font-bold tracking-tight text-neutral-900 dark:text-neutral-200",
                    tier.featured && "text-black dark:text-white"
                  )}
                >
                  {active === "monthly" ? tier.priceMonthly : tier.priceYearly}
                </motion.span>
              </p>
              <p
                className={cn(
                  "mt-6 h-12 text-sm leading-7 text-neutral-600 md:h-12 xl:h-12 dark:text-neutral-300",
                  tier.featured && "text-neutral-600 dark:text-neutral-300"
                )}
              >
                {tier.description}
              </p>
              <ul
                role="list"
                className={cn(
                  "mt-2 space-y-3 text-sm leading-6 text-neutral-600 sm:mt-4 dark:text-neutral-300",
                  tier.featured && "text-neutral-600 dark:text-neutral-300"
                )}
              >
                {tier.features.map((feature) => (
                  <li key={feature} className="flex gap-x-3">
                    <IconCircleCheckFilled
                      className={cn(
                        "h-6 w-5 flex-none text-neutral-700 dark:text-neutral-400",
                        tier.featured && "text-black dark:text-white"
                      )}
                      aria-hidden="true"
                    />
                    {feature}
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <button
                onClick={tier.onClick}
                aria-describedby={tier.id}
                className={cn(
                  "mt-8 block w-full rounded-lg bg-gradient-to-b from-indigo-500 to-indigo-600 px-3.5 py-2.5 text-center text-sm font-semibold text-white shadow-[0px_2px_0px_0px_rgba(255,255,255,0.5)_inset] transition duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 sm:mt-10"
                )}
              >
                {tier.cta}
              </button>
            </div>
          </div>
        ))}
      </div>
      <AddOn />
    </div>
  );
}

const AddOn = () => {
  return (
    <section className="relative z-20 mx-auto my-20 grid w-full max-w-6xl grid-cols-1 justify-start bg-gradient-to-br from-gray-100 to-white md:my-40 md:grid-cols-3 dark:from-neutral-900 dark:to-neutral-950">
      <GridLineHorizontal className="top-0" offset="200px" />
      <GridLineHorizontal className="bottom-0 top-auto" offset="200px" />
      <GridLineVertical className="left-4" offset="80px" />
      <GridLineVertical className="left-auto right-4" offset="80px" />
      <div className="p-8 md:col-span-2 md:p-14">
        <h2 className="text-left text-xl font-medium tracking-tight text-neutral-500 md:text-3xl dark:text-neutral-200">
          Buy for your team of{" "}
          <span className="font-bold text-black dark:text-white">
            10 people{" "}
          </span>
          and get
          <span className="font-bold text-indigo-500 dark:text-indigo-500">
            {" "}
            pro upgrade absolutely free
          </span>
          .
        </h2>

        <div className="flex flex-col items-start sm:flex-row sm:items-center sm:gap-4">
          <button className="group mt-8 flex items-center space-x-2 rounded-lg bg-gradient-to-b from-indigo-500 to-indigo-700 px-4 py-2 text-base text-white shadow-[0px_2px_0px_0px_rgba(255,255,255,0.3)_inset]">
            <span>Buy now</span>
          </button>
          <button className="group mt-8 flex items-center space-x-2 rounded-lg border border-neutral-200 px-4 py-2 text-base text-black shadow-[0px_2px_0px_0px_rgba(255,255,255,0.3)_inset] dark:border-neutral-800 dark:text-white">
            <span>Talk to us</span>
            <IconMessageCircleQuestion className="mt-0.5 h-3 w-3 stroke-[1px] text-black transition-transform duration-200 group-hover:translate-x-1 dark:text-white" />
          </button>
        </div>
      </div>
      <div className="border-t border-dashed p-8 md:border-l md:border-t-0 md:p-14">
        <p className="text-base text-neutral-700 dark:text-neutral-200">
          &quot;This is the best product ever when it comes to shipping. Ten on
          ten recommended. I just can&apos;t wait to see what happens with this
          product.&quot;
        </p>
        <div className="mt-4 flex flex-col items-start gap-1 text-sm">
          <p className="font-bold text-neutral-800 dark:text-neutral-200">
            Michael Scarn
          </p>
          <p className="text-neutral-500 dark:text-neutral-400">
            Side projects builder
          </p>
        </div>
      </div>
    </section>
  );
};

const GridLineHorizontal = ({
  className,
  offset,
}: {
  className?: string;
  offset?: string;
}) => {
  return (
    <div
      style={
        {
          "--background": "#ffffff",
          "--color": "rgba(0, 0, 0, 0.2)",
          "--height": "1px",
          "--width": "5px",
          "--fade-stop": "90%",
          "--offset": offset || "200px", //-100px if you want to keep the line inside
          "--color-dark": "rgba(255, 255, 255, 0.2)",
          maskComposite: "exclude",
        } as React.CSSProperties
      }
      className={cn(
        "absolute left-[calc(var(--offset)/2*-1)] h-[var(--height)] w-[calc(100%+var(--offset))]",
        "bg-[linear-gradient(to_right,var(--color),var(--color)_50%,transparent_0,transparent)]",
        "[background-size:var(--width)_var(--height)]",
        "[mask:linear-gradient(to_left,var(--background)_var(--fade-stop),transparent),_linear-gradient(to_right,var(--background)_var(--fade-stop),transparent),_linear-gradient(black,black)]",
        "[mask-composite:exclude]",
        "z-30",
        "dark:bg-[linear-gradient(to_right,var(--color-dark),var(--color-dark)_50%,transparent_0,transparent)]",
        className
      )}
    ></div>
  );
};

const GridLineVertical = ({
  className,
  offset,
}: {
  className?: string;
  offset?: string;
}) => {
  return (
    <div
      style={
        {
          "--background": "#ffffff",
          "--color": "rgba(0, 0, 0, 0.2)",
          "--height": "5px",
          "--width": "1px",
          "--fade-stop": "90%",
          "--offset": offset || "150px", //-100px if you want to keep the line inside
          "--color-dark": "rgba(255, 255, 255, 0.2)",
          maskComposite: "exclude",
        } as React.CSSProperties
      }
      className={cn(
        "absolute top-[calc(var(--offset)/2*-1)] h-[calc(100%+var(--offset))] w-[var(--width)]",
        "bg-[linear-gradient(to_bottom,var(--color),var(--color)_50%,transparent_0,transparent)]",
        "[background-size:var(--width)_var(--height)]",
        "[mask:linear-gradient(to_top,var(--background)_var(--fade-stop),transparent),_linear-gradient(to_bottom,var(--background)_var(--fade-stop),transparent),_linear-gradient(black,black)]",
        "[mask-composite:exclude]",
        "z-30",
        "dark:bg-[linear-gradient(to_bottom,var(--color-dark),var(--color-dark)_50%,transparent_0,transparent)]",
        className
      )}
    ></div>
  );
};

export type Tier = {
  name: string;
  id: string;
  href: string;
  priceMonthly: string;
  priceYearly: string;
  description: string;
  features: string[];
  featured: boolean;
  cta: string;
  onClick: () => void;
};

export const tiers: Tier[] = [
  {
    name: "Hobby",
    id: "tier-hobby",
    href: "#",
    priceMonthly: "$4/mo",
    priceYearly: "$30/yr",
    description: "Best for occasional listeners",
    features: [
      "Stream music on 1 device",
      "Access to basic music library",
      "Standard audio quality",
      "Ad-supported listening",
      "Create and share playlists",
    ],
    featured: false,
    cta: "Get Started",
    onClick: () => {},
  },
  {
    name: "Professional",
    id: "tier-professional",
    href: "#",
    priceMonthly: "$8/mo",
    priceYearly: "$60/yr",
    description: "Best for regular listeners",
    features: [
      "Stream music on up to 3 devices simultaneously",
      "Access to premium music library",
      "High-definition audio quality",
      "Ad-free listening experience",
      "Download music for offline listening",
      "Create and share unlimited playlists",
      "Access to exclusive content and early releases",
    ],
    featured: true,
    cta: "Get Started",
    onClick: () => {},
  },

  {
    name: "Enterprise",
    id: "tier-enterprise",
    href: "#",
    priceMonthly: "Contact Us",
    priceYearly: "Contact Us",
    description: "Best for big artists",
    features: [
      "Stream music on unlimited devices",
      "Access to entire music library",
      "Ultra high-definition audio quality",
      "Ad-free listening experience",
      "Download unlimited music for offline listening",
      "Create and share unlimited playlists",
      "Access to exclusive content and early releases",
      "Priority customer support",
    ],
    featured: false,
    cta: "Contact Us",
    onClick: () => {},
  },
];

export const Icon = ({ className, ...rest }: React.SVGProps<SVGSVGElement>) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth="1"
      stroke="currentColor"
      className={cn(
        "h-4 w-4 text-neutral-400 md:h-8 md:w-8 dark:text-neutral-600",
        className
      )}
      {...rest}
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v12m6-6H6" />
    </svg>
  );
};
