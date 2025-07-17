"use client";
import { cn } from "@/lib/utils";
import { IconArrowsUp, IconServer, IconTimeline } from "@tabler/icons-react";
import React from "react";
import { useId } from "react";

export function StatsWithGridBackground() {
  const items = [
    {
      title: "Years in the business",
      description: "Years in the business where we helped companies grow.",
      icon: IconTimeline,
      value: "10+",
    },
    {
      title: "People Served",
      description: "Over 100k people have used our services.",
      icon: IconServer,
      value: "100k+",
    },
    {
      title: "Response Time",
      description: "We are always available to help you with your needs.",
      icon: IconArrowsUp,
      value: "100%",
    },
  ];
  return (
    <div className="py-20">
      <div className="mx-auto max-w-7xl border border-neutral-200 dark:border-neutral-800">
        <div className="grid grid-cols-1 md:grid-cols-3">
          {items.map((item, index) => (
            <div
              key={"card" + index}
              className={cn(
                "group/card relative overflow-hidden p-10",
                index !== items.length - 1 &&
                  "border-b border-neutral-200 dark:border-neutral-800 md:border-b-0 md:border-r"
              )}
            >
              <Grid size={20} />
              <EdgeElement />

              <div className="flex items-center gap-2">
                <IconContainer>
                  <item.icon className="text-white" />
                </IconContainer>
                <p className="text-3xl font-bold text-neutral-700 dark:text-neutral-200">
                  {item.value}
                </p>
              </div>
              <p className="text-balance text-balance mt-4 text-base text-neutral-600 dark:text-neutral-300">
                {item.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

const EdgeElement = () => {
  return (
    <div className="absolute right-0 top-0 h-10 w-10 overflow-hidden border-b border-l bg-white shadow-[-3px_4px_9px_0px_rgba(0,0,0,0.14)] transition duration-200 group-hover/card:-translate-y-14 group-hover/card:translate-x-14 dark:border-neutral-800 dark:bg-neutral-900 dark:shadow-[-3px_4px_9px_0px_rgba(255,255,255,0.2)]">
      <div className="absolute left-0 top-0 h-[1px] w-[141%] origin-top-left rotate-45 bg-neutral-100 dark:bg-neutral-800" />
    </div>
  );
};
const IconContainer = ({ children }: { children: React.ReactNode }) => {
  return (
    <div className="flex h-16 w-16 items-center justify-center rounded-xl bg-gradient-to-b from-neutral-200 to-white to-[50%] p-1 dark:from-neutral-800 dark:to-black">
      <div className="flex h-full w-full items-center justify-center rounded-lg bg-gradient-to-b from-[#5D5D5D] to-black dark:to-neutral-900">
        {children}
      </div>
    </div>
  );
};

export const Grid = ({
  pattern,
  size,
}: {
  pattern?: number[][];
  size: number;
}) => {
  const p = pattern ?? [
    [Math.floor(Math.random() * 4) + 7, Math.floor(Math.random() * 6) + 1],
    [Math.floor(Math.random() * 4) + 7, Math.floor(Math.random() * 6) + 1],
    [Math.floor(Math.random() * 4) + 7, Math.floor(Math.random() * 6) + 1],
    [Math.floor(Math.random() * 4) + 7, Math.floor(Math.random() * 6) + 1],
    [Math.floor(Math.random() * 4) + 7, Math.floor(Math.random() * 6) + 1],
  ];
  return (
    <div className="pointer-events-none absolute left-1/2 top-0 -ml-20 -mt-2 h-full w-full [mask-image:linear-gradient(white,transparent)]">
      <div className="absolute inset-0 bg-gradient-to-r from-zinc-100/30 to-zinc-300/30 opacity-100 [mask-image:radial-gradient(farthest-side_at_top,white,transparent)] dark:from-zinc-900/30 dark:to-zinc-900/30">
        <GridPattern
          width={size ?? 20}
          height={size ?? 20}
          x="-12"
          y="4"
          squares={p}
          className="absolute inset-0 h-full w-full fill-black/10 stroke-black/10 mix-blend-overlay dark:fill-white/10 dark:stroke-white/10"
        />
      </div>
    </div>
  );
};

export function GridPattern({
  width,
  height,
  x,
  y,
  squares,
  ...props
}: {
  width: number;
  height: number;
  x: number | string;
  y: number | string;
  squares: number[][];
} & React.SVGProps<SVGSVGElement>) {
  const patternId = useId();

  return (
    <svg aria-hidden="true" {...props}>
      <defs>
        <pattern
          id={patternId}
          width={width}
          height={height}
          patternUnits="userSpaceOnUse"
          x={x}
          y={y}
        >
          <path d={`M.5 ${height}V.5H${width}`} fill="none" />
        </pattern>
      </defs>
      <rect
        width="100%"
        height="100%"
        strokeWidth={0}
        fill={`url(#${patternId})`}
      />
      {squares && (
        <svg x={x} y={y} className="overflow-visible">
          {squares.map(([x, y]) => (
            <rect
              strokeWidth="0"
              key={`${x}-${y}`}
              width={width + 1}
              height={height + 1}
              x={x * width}
              y={y * height}
            />
          ))}
        </svg>
      )}
    </svg>
  );
}
