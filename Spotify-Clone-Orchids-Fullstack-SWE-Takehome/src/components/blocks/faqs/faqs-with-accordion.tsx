"use client";
import React, { useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { IconMinus, IconPlus } from "@tabler/icons-react";
import { cn } from "@/lib/utils";

const FAQs = [
  {
    question: "What is the purpose of this website?",
    answer:
      "This website is a place to help you find the best products and services in the world.",
  },
  {
    question: "How do I contact support?",
    answer:
      "You can contact support by email at support@example.com or by phone at 123-456-7890.",
  },
  {
    question: "How do I find the best products?",
    answer:
      "You can find the best products by searching for them on the search page or by browsing the categories.",
  },
  {
    question: "Can I return a product?",
    answer:
      "Yes, you can return a product within 30 days of purchase. Please refer to our return policy for more details.",
  },
  {
    question: "Do you offer international shipping?",
    answer:
      "Yes, we offer international shipping to most countries. Shipping fees and delivery times may vary depending on the destination.",
  },
  {
    question: "How can I track my order?",
    answer:
      "You can track your order by logging into your account and visiting the order history page. You will also receive a tracking number via email once your order has shipped.",
  },
];
export function FrequentlyAskedQuestionsAccordion() {
  const [open, setOpen] = useState<string | null>(null);
  return (
    <div className="mx-auto grid max-w-6xl grid-cols-1 gap-4 px-4 py-20 md:grid-cols-2 md:px-8 md:py-40">
      <h2 className="text-center text-4xl font-bold tracking-tight text-neutral-600 md:text-left md:text-6xl dark:text-neutral-50">
        Frequently asked questions
      </h2>
      <div className="divide-y divide-neutral-200 dark:divide-neutral-800">
        {FAQs.map((faq, index) => (
          <FAQItem
            key={index}
            question={faq.question}
            answer={faq.answer}
            open={open}
            setOpen={setOpen}
          />
        ))}
      </div>
    </div>
  );
}

const FAQItem = ({
  question,
  answer,
  setOpen,
  open,
}: {
  question: string;
  answer: string;
  open: string | null;
  setOpen: (open: string | null) => void;
}) => {
  const isOpen = open === question;

  return (
    <div
      className="cursor-pointer py-4"
      onClick={() => {
        if (isOpen) {
          setOpen(null);
        } else {
          setOpen(question);
        }
      }}
    >
      <div className="flex items-start">
        <div className="relative mr-4 mt-1 h-6 w-6 flex-shrink-0">
          <IconPlus
            className={cn(
              "absolute inset-0 h-6 w-6 transform text-blue-500 transition-all duration-200",
              isOpen && "rotate-90 scale-0"
            )}
          />
          <IconMinus
            className={cn(
              "absolute inset-0 h-6 w-6 rotate-90 scale-0 transform text-blue-500 transition-all duration-200",
              isOpen && "rotate-0 scale-100"
            )}
          />
        </div>
        <div>
          <h3 className="text-lg font-medium text-neutral-700 dark:text-neutral-200">
            {question}
          </h3>
          <AnimatePresence mode="wait">
            {isOpen && (
              <motion.div
                initial={{ height: 0 }}
                animate={{ height: "auto" }}
                exit={{ height: 0 }}
                transition={{ duration: 0.2, ease: "easeOut" }}
                className="overflow-hidden text-neutral-500 dark:text-neutral-400"
              >
                <p>{answer}</p>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
};
