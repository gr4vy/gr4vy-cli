import {EmbedParams} from '@gr4vy/sdk'

export const parseEmbedParams = (amount: number, currency: string, argv: string[]): EmbedParams => {
  const params: Record<string, any> = {
    amount, 
    currency
  }

  argv
    .map((arg) => arg.split('='))
    .filter(([key, value]) => key && value)
    .forEach(([key, value]) => (params[key] = value))

  return params as EmbedParams
}
