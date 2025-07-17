"use client";
import React from "react";

const FAQs = [
  {
    question: "What is the purpose of this website?",
    answer:
      "This website is a place to help you find the best products and services in the world. We aim to provide comprehensive reviews and comparisons to help you make informed decisions.",
  },
  {
    question: "How do I contact support?",
    answer:
      "You can contact support by email at support@example.com or by phone at 123-456-7890. ",
  },
  {
    question: "How do I find the best products?",
    answer:
      "You can find the best products by searching for them on the search page or by browsing the categories. Our curated lists and expert reviews will guide you to the top choices.",
  },
  {
    question: "Can I return a product?",
    answer:
      "Yes, you can return a product within 30 days of purchase. Please refer to our return policy for more details. Ensure the product is in its original condition and packaging.",
  },
  {
    question: "Do you offer international shipping?",
    answer:
      "Yes, we offer international shipping to most countries. Shipping fees and delivery times may vary depending on the destination. ",
  },
  {
    question: "How can I track my order?",
    answer:
      "You can track your order by logging into your account and visiting the order history page. You will also receive a tracking number via email once your order has shipped. Stay updated with real-time tracking.",
  },
  {
    question: "What payment methods do you accept?",
    answer:
      "We accept various payment methods including credit cards, debit cards, PayPal, and bank transfers. Choose the most convenient option for you during checkout.",
  },
  {
    question: "How do I reset my password?",
    answer:
      "You can reset your password by clicking on the 'Forgot Password' link on the login page and following the instructions. A password reset link will be sent to your registered email address.",
  },
  {
    question: "Can I change my shipping address after placing an order?",
    answer:
      "Yes, you can change your shipping address before the order is shipped. Please contact our support team for assistance. Provide the new address details promptly to avoid any delays.",
  },
  {
    question: "What is your privacy policy?",
    answer:
      "Our privacy policy outlines how we collect, use, and protect your personal information. You can read it on our privacy policy page. We are committed to safeguarding your privacy.",
  },
  {
    question: "How do I leave a review for a product?",
    answer:
      "You can leave a review for a product by logging into your account, navigating to the product page, and clicking on the 'Write a Review' button. Share your experience to help others make informed decisions.",
  },
];
export function FAQsWithGrid() {
  const columns = 3;
  const faqsGrid: { question: string; answer: string }[][] = Array.from(
    { length: columns },
    () => []
  );
  FAQs.forEach((faq, index) => {
    faqsGrid[index % columns].push(faq);
  });
  return (
    <div className="mx-auto grid max-w-7xl gap-4 px-4 py-20 md:px-8 md:py-40">
      <h2 className="text-left text-4xl font-medium tracking-tight text-neutral-600 dark:text-neutral-50 md:text-5xl">
        Frequently asked questions
      </h2>
      <p className="max-w-lg text-left text-base text-neutral-600 dark:text-neutral-50">
        We are here to help you with any questions you may have. If you
        don&apos;t find what you need, please contact us at{" "}
        <a
          href="mailto:support@example.com"
          className="text-blue-500 underline"
        >
          support@example.com
        </a>
      </p>
      <div className="mt-10 grid w-full grid-cols-1 items-start gap-4 md:grid-cols-3">
        {faqsGrid.map((faqs, rowIndex) => (
          <div key={rowIndex} className="grid grid-cols-1 items-start gap-4">
            {faqs.map((faq, index) => (
              <FAQItem
                key={index}
                question={faq.question}
                answer={faq.answer}
              />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

const FAQItem = ({
  question,
  answer,
}: {
  question: string;
  answer: string;
}) => {
  return (
    <div className="cursor-pointer rounded-2xl bg-white p-8 shadow-[0_1px_1px_rgba(0,0,0,0.05),0_4px_6px_rgba(34,42,53,0.04),0_24px_68px_rgba(47,48,55,0.05),0_2px_3px_rgba(0,0,0,0.04)] dark:bg-neutral-900 dark:shadow-[2px_4px_16px_0px_rgba(248,248,248,0.06)_inset]">
      <h3 className="text-lg font-medium text-neutral-700 dark:text-neutral-200">
        {question}
      </h3>

      <p className="mt-2 text-sm text-neutral-500 dark:text-neutral-400 lg:text-base">
        {answer}
      </p>
    </div>
  );
};
