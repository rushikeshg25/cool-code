"use client";
import React from "react";
import { IconCircleCheckFilled } from "@tabler/icons-react";
import { cn } from "@/lib/utils";
import { useState } from "react";
import { motion } from "motion/react";
import { Outfit } from "next/font/google";

const outfit = Outfit({ subsets: ["latin"], weight: ["400", "500", "700"] });
export function PricingWithSwitch() {
  return (
    <div className="relative mx-auto flex max-w-7xl flex-col items-center justify-between py-10 md:py-20">
      <div className="relative">
        <h2
          className={cn(
            "text-center text-xl font-bold text-black md:text-4xl dark:text-white",
            outfit.className
          )}
        >
          Simple pricing for your ease
        </h2>
        <p
          className={cn(
            "mx-auto mt-4 max-w-2xl text-center text-sm font-normal text-neutral-700 md:text-base dark:text-neutral-400",
            outfit.className
          )}
        >
          Every AI offers a wide range of services. You can choose the one that
          suits your needs. Select from your favourite plan and get started
          instantly.
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
      <div className="mx-auto mb-12 mt-10 flex w-fit items-center justify-center overflow-hidden rounded-md bg-neutral-100 dark:bg-neutral-800">
        {tabs.map((tab) => (
          <button
            key={tab.value}
            className={cn(
              "relative rounded-md p-4 text-sm font-medium text-gray-500 dark:text-neutral-400",
              active === tab.value ? "text-white dark:text-black" : ""
            )}
            onClick={() => setActive(tab.value)}
          >
            {active === tab.value && (
              <motion.span
                layoutId="moving-div"
                transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                className="absolute inset-0 bg-black dark:bg-white"
              />
            )}
            <span className="relative z-10">{tab.name}</span>
          </button>
        ))}
      </div>
      <div className="relative z-20 mx-auto mt-4 grid grid-cols-1 items-center gap-4 md:mt-20 md:grid-cols-2 xl:grid-cols-4">
        {tiers.map((tier, tierIdx) => (
          <div
            key={tier.id}
            className={cn(
              tier.featured
                ? "relative bg-[radial-gradient(164.75%_100%_at_50%_0%,#334155_0%,#0F172A_48.73%)] shadow-2xl"
                : "bg-white dark:bg-black",
              "flex h-full flex-col justify-between rounded-lg px-6 py-8 sm:mx-8 lg:mx-0"
            )}
          >
            <div className="">
              <h3
                id={tier.id}
                className={cn(
                  tier.featured
                    ? "text-white"
                    : "text-neutral-700 dark:text-neutral-400",
                  "text-base font-semibold leading-7"
                )}
              >
                {tier.name}
              </h3>
              <p className="mt-4">
                <motion.span
                  initial={{ y: 20, opacity: 0, filter: "blur(10px)" }}
                  animate={{ y: 0, opacity: 1, filter: "blur(0px)" }}
                  transition={{
                    duration: 0.2,
                    ease: "easeOut",
                    delay: 0.1 * tierIdx,
                  }}
                  key={active}
                  className={cn(
                    "inline-block text-4xl font-bold tracking-tight",
                    tier.featured
                      ? "text-white"
                      : "text-neutral-900 dark:text-neutral-200"
                  )}
                >
                  {active === "monthly" ? tier.priceMonthly : tier.priceYearly}
                </motion.span>
              </p>
              <p
                className={cn(
                  tier.featured
                    ? "text-neutral-300"
                    : "text-neutral-600 dark:text-neutral-300",
                  "mt-6 h-12 text-sm leading-7 md:h-12 xl:h-12"
                )}
              >
                {tier.description}
              </p>
              <ul
                role="list"
                className={cn(
                  tier.featured
                    ? "text-neutral-300"
                    : "text-neutral-600 dark:text-neutral-300",
                  "mt-8 space-y-3 text-sm leading-6 sm:mt-10"
                )}
              >
                {tier.features.map((feature) => (
                  <li key={feature} className="flex gap-x-3">
                    <IconCircleCheckFilled
                      className={cn(
                        tier.featured
                          ? "text-white"
                          : "text-neutral-700 dark:text-neutral-400",
                        "h-6 w-5 flex-none"
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
                  "mt-8 block w-full rounded-full bg-black px-3.5 py-2.5 text-center text-sm font-semibold text-white ring-black ring-offset-2 transition duration-200 hover:ring-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 sm:mt-10 dark:bg-white dark:text-black dark:ring-white dark:ring-offset-2",
                  tier.featured
                    ? "bg-white text-black shadow-sm hover:bg-white/90 focus-visible:outline-white dark:bg-white"
                    : ""
                )}
              >
                {tier.cta}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

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
    name: "Starter",
    id: "tier-starter",
    href: "#",
    priceMonthly: "$4/mo",
    priceYearly: "$30/yr",
    description: "Best for creators starting out",
    features: [
      "Manage up to 2 channels",
      "Basic video upload support",
      "Email support within 48 hours",
      "Access to community forum",
      "Monthly performance reports",
    ],
    featured: false,
    cta: "Browse Starter",
    onClick: () => {},
  },
  {
    name: "Medium",
    id: "tier-medium",
    href: "#",
    priceMonthly: "$8/mo",
    priceYearly: "$60/yr",
    description: "Perfect for creaters between 10k - 100k",
    features: [
      "Everything in Starter, plus",
      "Manage up to 5 channels",
      "Priority video upload support",
      "Email support within 24 hours",
      "Access to exclusive webinars",
      "Monthly strategy sessions",
      "Advanced analytics reports",
    ],
    featured: false,
    cta: "Buy Now",
    onClick: () => {},
  },
  {
    name: "Influencer",
    id: "tier-influencer",
    href: "#",
    priceMonthly: "$12/mo",
    priceYearly: "$100/yr",
    description: "Perfect for creaters between 100k - 1m",
    features: [
      "Everything in Starter, plus",
      "Manage up to 10 channels",
      "Priority video upload support",
      "Email support within 12 hours",
      "Access to private creator community",
      "Monthly strategy sessions",
      "Advanced analytics and insights",
    ],
    featured: true,
    cta: "Buy Now",
    onClick: () => {},
  },

  {
    name: "Celebrity",
    id: "tier-celebrity",
    href: "#",
    priceMonthly: "Contact Us",
    priceYearly: "Contact Us",
    description: "Perfect for creaters between 1m - 100m",
    features: [
      "Everything in Influencer, plus",
      "Manage unlimited channels",
      "24/7 priority support",
      "Access to VIP creator community",
      "Weekly strategy sessions",
      "Comprehensive analytics and insights",
      "Custom branding and design services",
    ],
    featured: false,
    cta: "Contact Us",
    onClick: () => {},
  },
];
