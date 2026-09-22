export interface CardToIbanData {
  requestId: string
  trackId: string
  bankName: string
  englishBankName: string
  iban: string
  deposit: string
  depositDescription: string
  depositOwners: string
  depositStatus: string
}

export interface CardToIbanResponse {
  data: CardToIbanData
}

export interface CardToIbanErrorResponse {
  error: {
    code: string
    message: string
    requestId?: string
    validationCode?: string
  }
}
