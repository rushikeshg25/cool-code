import { CheckIcon } from '@heroicons/react/20/solid'

const tiers = [
  {
    id: 'freelancer',
    name: 'Freelancer',
    price: { monthly: '$19', annually: '$199' },
    description: 'The essentials to provide your best work for clients.',
    features: ['5 products', 'Up to 1,000 subscribers', 'Basic analytics', '48-hour support response time'],
    featured: false,
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
    featured: true,
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    price: { monthly: '$59', annually: '$599' },
    description: 'Dedicated support and infrastructure for your company.',
    features: [
      'Unlimited products',
      'Unlimited subscribers',
      'Advanced analytics',
      '1-hour, dedicated support response time',
      'Marketing automations',
      'Custom reporting tools',
    ],
    featured: false,
  },
]

export default function ThreeTiersWithToggleOnDark() {
  return (
    <form className="group/tiers bg-gray-900 py-24 sm:py-32">
      <div className="mx-auto max-w-7xl px-6 lg:px-8">
        <div className="mx-auto max-w-4xl text-center">
          <h2 className="text-base/7 font-semibold text-indigo-400">Pricing</h2>
          <p className="mt-2 text-5xl font-semibold tracking-tight text-balance text-white sm:text-6xl">
            Pricing that grows with you
          </p>
        </div>
        <p className="mx-auto mt-6 max-w-2xl text-center text-lg font-medium text-pretty text-gray-400 sm:text-xl/8">
          Choose an affordable plan that’s packed with the best features for engaging your audience, creating customer
          loyalty, and driving sales.
        </p>
        <div className="mt-16 flex justify-center">
          <fieldset aria-label="Payment frequency">
            <div className="grid grid-cols-2 gap-x-1 rounded-full bg-white/5 p-1 text-center text-xs/5 font-semibold text-white">
              <label className="group relative rounded-full px-2.5 py-1 has-checked:bg-indigo-500">
                <input
                  defaultValue="monthly"
                  defaultChecked
                  name="frequency"
                  type="radio"
                  className="absolute inset-0 appearance-none rounded-full"
                />
                <span className="text-white">Monthly</span>
              </label>
              <label className="group relative rounded-full px-2.5 py-1 has-checked:bg-indigo-500">
                <input
                  defaultValue="annually"
                  name="frequency"
                  type="radio"
                  className="absolute inset-0 appearance-none rounded-full"
                />
                <span className="text-white">Annually</span>
              </label>
            </div>
          </fieldset>
        </div>
        <div className="isolate mx-auto mt-10 grid max-w-md grid-cols-1 gap-8 lg:mx-0 lg:max-w-none lg:grid-cols-3">
          {tiers.map((tier) => (
            <div
              key={tier.id}
              data-featured={tier.featured ? 'true' : undefined}
              className="group/tier rounded-3xl p-8 ring-1 ring-white/10 data-featured:bg-white/5 data-featured:ring-2 data-featured:ring-indigo-500 xl:p-10"
            >
              <div className="flex items-center justify-between gap-x-4">
                <h3 id={`tier-${tier.id}`} className="text-lg/8 font-semibold text-white">
                  {tier.name}
                </h3>
                <p className="rounded-full bg-indigo-500 px-2.5 py-1 text-xs/5 font-semibold text-white group-not-data-featured/tier:hidden">
                  Most popular
                </p>
              </div>
              <p className="mt-4 text-sm/6 text-gray-300">{tier.description}</p>
              <p className="mt-6 flex items-baseline gap-x-1 group-not-has-[[name=frequency][value=monthly]:checked]/tiers:hidden">
                <span className="text-4xl font-semibold tracking-tight text-white">{tier.price.monthly}</span>
                <span className="text-sm/6 font-semibold text-gray-300">/month</span>
              </p>
              <p className="mt-6 flex items-baseline gap-x-1 group-not-has-[[name=frequency][value=annually]:checked]/tiers:hidden">
                <span className="text-4xl font-semibold tracking-tight text-white">{tier.price.annually}</span>
                <span className="text-sm/6 font-semibold text-gray-300">/year</span>
              </p>
              <button
                value={tier.id}
                name="tier"
                type="submit"
                aria-describedby={`tier-${tier.id}`}
                className="mt-6 block w-full rounded-md bg-white/10 px-3 py-2 text-center text-sm/6 font-semibold text-white group-data-featured/tier:bg-indigo-500 group-data-featured/tier:text-white group-data-featured/tier:shadow-xs hover:bg-white/20 group-data-featured/tier:hover:bg-indigo-400 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white group-data-featured/tier:focus-visible:outline-indigo-500"
              >
                Buy plan
              </button>
              <ul role="list" className="mt-8 space-y-3 text-sm/6 text-gray-300 xl:mt-10">
                {tier.features.map((feature) => (
                  <li key={feature} className="flex gap-x-3">
                    <CheckIcon aria-hidden="true" className="h-6 w-5 flex-none text-white" />
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
