import type {
  CardToIbanData,
  CardToIbanErrorResponse,
  CardToIbanResponse,
} from '~/types/api'

const persianDigits = '۰۱۲۳۴۵۶۷۸۹'
const arabicDigits = '٠١٢٣٤٥٦٧٨٩'

export function normalizeCardNumber(value: string): string {
  return value
    .split('')
    .map((character) => {
      const persianIndex = persianDigits.indexOf(character)
      if (persianIndex >= 0) return String(persianIndex)

      const arabicIndex = arabicDigits.indexOf(character)
      if (arabicIndex >= 0) return String(arabicIndex)

      return character
    })
    .join('')
    .replace(/[\s-]/g, '')
}

export function formatCardNumber(value: string): string {
  const normalized = normalizeCardNumber(value).replace(/\D/g, '').slice(0, 16)
  return normalized.replace(/(.{4})/g, '$1 ').trim()
}

export function validateCardNumber(value: string): string | null {
  const normalized = normalizeCardNumber(value)

  if (!normalized) return 'شماره کارت را وارد کنید.'
  if (!/^\d+$/.test(normalized)) return 'شماره کارت فقط باید شامل ارقام باشد.'
  if (normalized.length !== 16) return 'شماره کارت باید ۱۶ رقم باشد.'
  if (/^(\d)\1{15}$/.test(normalized)) return 'شماره کارت واردشده معتبر نیست.'

  let sum = 0
  for (let index = 0; index < normalized.length; index += 1) {
    let digit = Number(normalized[index])
    if (index % 2 === 0) {
      digit *= 2
      if (digit > 9) digit -= 9
    }
    sum += digit
  }

  return sum % 10 === 0 ? null : 'ساختار شماره کارت معتبر نیست.'
}

export function formatIban(value: string): string {
  return value.replace(/\s/g, '').replace(/(.{4})/g, '$1 ').trim()
}

export function useCardToIban() {
  const config = useRuntimeConfig()
  const result = ref<CardToIbanData | null>(null)
  const errorMessage = ref('')
  const errorCode = ref('')
  const isLoading = ref(false)

  async function convert(card: string) {
    const validationError = validateCardNumber(card)
    if (validationError) {
      errorMessage.value = validationError
      errorCode.value = 'validation_error'
      result.value = null
      return false
    }

    isLoading.value = true
    errorMessage.value = ''
    errorCode.value = ''
    result.value = null

    try {
      const response = await $fetch<CardToIbanResponse>(
        `${config.public.apiBaseUrl}/card-to-iban`,
        {
          method: 'POST',
          body: { card: normalizeCardNumber(card) },
        },
      )
      result.value = response.data
      return true
    } catch (error: unknown) {
      const fetchError = error as { data?: CardToIbanErrorResponse }
      errorMessage.value =
        fetchError.data?.error?.message || 'ارتباط با سرویس استعلام برقرار نشد.'
      errorCode.value = fetchError.data?.error?.code || 'network_error'
      return false
    } finally {
      isLoading.value = false
    }
  }

  function reset() {
    result.value = null
    errorMessage.value = ''
    errorCode.value = ''
  }

  return {
    result,
    errorMessage,
    errorCode,
    isLoading,
    convert,
    reset,
  }
}
