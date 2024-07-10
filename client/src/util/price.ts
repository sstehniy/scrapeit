export const getPriceNumber = (price: string) => {
  // handle different cases, e.g. "1000.00" or "1,000.00"
  if (price.includes(",")) {
    return parseFloat(price.replace(/,/g, ""));
  }
  return parseFloat(price);
};
