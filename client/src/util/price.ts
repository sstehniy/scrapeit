export const getPriceNumber = (price: string): number => {
  if (price.includes(".") && price.includes(",")) {
    return parseFloat(price.replace(",", ".").replace(/\./g, ""));
  } else if (price.includes(",")) {
    return parseFloat(price.replace(",", ""));
  } else if (price.includes(".")) {
    return parseFloat(price.replace(/\./g, ""));
  }
  return parseFloat(price);
};
