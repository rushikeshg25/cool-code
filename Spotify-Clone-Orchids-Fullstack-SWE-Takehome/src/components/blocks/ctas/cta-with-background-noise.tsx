"use client";
import Image from "next/image";
import React from "react";
import { HiArrowRight } from "react-icons/hi2";

export function CTAWithBackgroundNoise() {
  return (
    <section className="w-full grid grid-cols-1 md:grid-cols-2 my-20 md:my-40 justify-start relative z-20 max-w-7xl mx-auto bg-gradient-to-br from-slate-800 dark:from-neutral-900 to-gray-900 sm:rounded-2xl">
      <div className="absolute -top-px right-10 md:right-60 bg-gradient-to-r from-transparent via-purple-500 h-px to-transparent w-1/2 z-30"></div>
      <div className="absolute -top-px right-10 md:right-40 bg-gradient-to-r from-transparent via-indigo-500 h-px to-transparent w-1/2 z-30"></div>
      <div className="absolute -top-px right-10 md:right-80 bg-gradient-to-r from-transparent via-sky-500 h-px to-transparent w-1/2 z-30"></div>
      <div
        className="absolute inset-0 w-full h-full opacity-10 bg-noise [mask-image:radial-gradient(#fff,transparent,75%)]"
        style={{
          backgroundImage: "url(/noise.webp)",
          backgroundSize: "30%",
        }}
      ></div>
      <div className="relative sm:rounded-2xl overflow-hidden px-6 md:px-8 ">
        <div
          className="pointer-events-none absolute inset-y-0 right-0 select-none overflow-hidden rounded-2xl"
          style={{
            mask: "radial-gradient(33.875rem 33.875rem at calc(100% - 8.9375rem) 0, white 3%, transparent 70%)",
          }}
        ></div>

        <div className="relative px-0 py-10 sm:px-10 sm:pt-20 sm:pb-10 lg:px-10">
          <h2 className="text-left text-balance  text-2xl md:text-3xl lg:text-5xl font-semibold tracking-[-0.015em] text-white">
            Ready to try out the product?
          </h2>
          <p className="mt-8 max-w-[26rem] text-left  text-base/6 text-neutral-200">
            Get instant access to our state of the art project and join the
            waitlist.
          </p>

          <button className="mt-8 flex space-x-2 items-center group text-base px-4 py-2 rounded-lg bg-gradient-to-b from-blue-500 to-blue-700 text-white shadow-[0px_2px_0px_0px_rgba(255,255,255,0.3)_inset]">
            <span>Join Waitlist</span>
            <HiArrowRight className="text-white group-hover:translate-x-1 stroke-[1px] h-3 w-3 mt-0.5 transition-transform duration-200" />
          </button>
        </div>
      </div>
      <div className="relative h-full flex gap-4 w-full overflow-hidden max-h-[200px] md:max-h-[400px] px-4">
        <Image
          src="https://assets.aceternity.com/pro/cta-1.jpg"
          alt="cta-1"
          width="300"
          height="500"
          className="h-full object-cover object-top rounded-lg md:rounded-none mt-4 md:mt-0"
        />
        <Image
          src="https://assets.aceternity.com/pro/cta-2.jpg"
          alt="cta-1"
          width="300"
          height="500"
          className="h-full object-cover object-top mt-10 rounded-lg"
        />
      </div>
      <div className="absolute -bottom-px right-10 md:right-60 bg-gradient-to-r from-transparent via-purple-500 h-px to-transparent w-1/2 z-30"></div>
      <div className="absolute -bottom-px right-10 md:right-40 bg-gradient-to-r from-transparent via-indigo-500 h-px to-transparent w-1/2 z-30"></div>
      <div className="absolute -bottom-px right-10 md:right-80 bg-gradient-to-r from-transparent via-sky-500 h-px to-transparent w-1/2 z-30"></div>
    </section>
  );
}
