import { CheckIcon } from '@heroicons/react/20/solid'

const tiers = [
  {
    id: 'freelancer',
    name: 'Freelancer',
    price: { monthly: '$19', annually: '$199' },
    description: 'The essentials to provide your best work for clients.',
    features: ['5 products', 'Up to 1,000 subscribers', 'Basic analytics', '48-hour support response time'],
    featured: false,
    cta: 'Buy plan',
  },
  {
    id: 'startup',
    name: 'Startup',
    price: { monthly: '$29', annually: '$299' },
    description: 'A plan that scales with your rapidly growing business.',
    features: [
      '25 products',
      'Up to 10,000 subscribers',
      'Advanced analytics',
      '24-hour support response time',
      'Marketing automations',
    ],
    featured: false,
    cta: 'Buy plan',
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    price: 'Custom',
    description: 'Dedicated support and infrastructure for your company.',
    features: [
      'Unlimited products',
      'Unlimited subscribers',
      'Advanced analytics',
      '1-hour, dedicated support response time',
      'Marketing automations',
      'Custom reporting tools',
    ],
    featured: true,
    cta: 'Contact sales',
  },
]

export default function ThreeTiersWithEmphasizedTier() {
  return (
    <form className="group/tiers bg-white py-24 sm:py-32">
      <div className="mx-auto max-w-7xl px-6 lg:px-8">
        <div className="mx-auto max-w-4xl text-center">
          <h2 className="text-base/7 font-semibold text-indigo-600">Pricing</h2>
          <p className="mt-2 text-5xl font-semibold tracking-tight text-balance text-gray-900 sm:text-6xl">
            Pricing that grows with you
          </p>
        </div>
        <p className="mx-auto mt-6 max-w-2xl text-center text-lg font-medium text-pretty text-gray-600 sm:text-xl/8">
          Choose an affordable plan that’s packed with the best features for engaging your audience, creating customer
          loyalty, and driving sales.
        </p>
        <div className="mt-16 flex justify-center">
          <fieldset aria-label="Payment frequency">
            <div className="grid grid-cols-2 gap-x-1 rounded-full p-1 text-center text-xs/5 font-semibold ring-1 ring-gray-200 ring-inset">
              <label className="group relative rounded-full px-2.5 py-1 has-checked:bg-indigo-600">
                <input
                  defaultValue="monthly"
                  defaultChecked
                  name="frequency"
                  type="radio"
                  className="absolute inset-0 appearance-none rounded-full"
                />
                <span className="text-gray-500 group-has-checked:text-white">Monthly</span>
              </label>
              <label className="group relative rounded-full px-2.5 py-1 has-checked:bg-indigo-600">
                <input
                  defaultValue="annually"
                  name="frequency"
                  type="radio"
                  className="absolute inset-0 appearance-none rounded-full"
                />
                <span className="text-gray-500 group-has-checked:text-white">Annually</span>
              </label>
            </div>
          </fieldset>
        </div>
        <div className="isolate mx-auto mt-10 grid max-w-md grid-cols-1 gap-8 lg:mx-0 lg:max-w-none lg:grid-cols-3">
          {tiers.map((tier) => (
            <div
              key={tier.id}
              data-featured={tier.featured ? 'true' : undefined}
              className="group/tier rounded-3xl p-8 ring-1 ring-gray-200 data-featured:bg-gray-900 data-featured:ring-gray-900 xl:p-10"
            >
              <h3
                id={`tier-${tier.id}`}
                className="text-lg/8 font-semibold text-gray-900 group-data-featured/tier:text-white"
              >
                {tier.name}
              </h3>
              <p className="mt-4 text-sm/6 text-gray-600 group-data-featured/tier:text-gray-300">{tier.description}</p>
              {typeof tier.price === 'string' ? (
                <p className="mt-6 text-4xl font-semibold tracking-tight text-gray-900 group-data-featured/tier:text-white">
                  {tier.price}
                </p>
              ) : (
                <>
                  <p className="mt-6 flex items-baseline gap-x-1 group-not-has-[[name=frequency][value=monthly]:checked]/tiers:hidden">
                    <span className="text-4xl font-semibold tracking-tight text-gray-900 group-data-featured/tier:text-white">
                      {tier.price.monthly}
                    </span>
                    <span className="text-sm/6 font-semibold text-gray-600 group-data-featured/tier:text-gray-300">
                      /month
                    </span>
                  </p>
                  <p className="mt-6 flex items-baseline gap-x-1 group-not-has-[[name=frequency][value=annually]:checked]/tiers:hidden">
                    <span className="text-4xl font-semibold tracking-tight text-gray-900 group-data-featured/tier:text-white">
                      {tier.price.annually}
                    </span>
                    <span className="text-sm/6 font-semibold text-gray-600 group-data-featured/tier:text-gray-300">
                      /year
                    </span>
                  </p>
                </>
              )}

              <button
                value={tier.id}
                name="tier"
                type="submit"
                aria-describedby={`tier-${tier.id}`}
                className="mt-6 block w-full rounded-md bg-indigo-600 px-3 py-2 text-center text-sm/6 font-semibold text-white shadow-xs group-data-featured/tier:bg-white/10 group-data-featured/tier:text-white hover:bg-indigo-500 group-data-featured/tier:hover:bg-white/20 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 group-data-featured/tier:focus-visible:outline-white"
              >
                {tier.cta}
              </button>
              <ul
                role="list"
                className="mt-8 space-y-3 text-sm/6 text-gray-600 group-data-featured/tier:text-gray-300 xl:mt-10"
              >
                {tier.features.map((feature) => (
                  <li key={feature} className="flex gap-x-3">
                    <CheckIcon
                      aria-hidden="true"
                      className="h-6 w-5 flex-none text-indigo-600 group-data-featured/tier:text-white"
                    />
                    {feature}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>
    </form>
  )
}
