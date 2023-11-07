import { z } from "zod";

const float = z.preprocess((val) => {
  if (typeof val === "string") {
    return parseFloat(val);
  }
  return val;
}, z.number());

const int = z.preprocess((val) => {
  if (typeof val === "string") {
    return parseInt(val);
  }
  return val;
}, z.number());

export const SoluteInputSchema = z.object({
  name: z.string(),
  molecular_weight: float
});

export type SoluteInput = z.infer<typeof SoluteInputSchema>

export const SolventInputSchema = z.object({
  name: z.string()
});

export type SolventInput = z.infer<typeof SolventInputSchema>

export const StockEntryInputSchema = z.object({
  name: z.string(),
  amount: float
});

export const SolutionInputSchema = z.object({
  name: z.string(),
  solutes: z.array(StockEntryInputSchema).optional(),
  solvents: z.array(StockEntryInputSchema).optional()
});

export type SolutionInput = z.infer<typeof SolutionInputSchema>

export const StockInputSchema = z.object({
  name: z.string(),
  grid: z.string(),
  pos: z.string(),
  volume: float,
  solution: SolutionInputSchema
});

export type StockInput = z.infer<typeof StockInputSchema>

export const ContainerFormatInputSchema = z.object({
  name: z.string(),
  n_wells: int,
  n_cols: int,
  n_rows: int,
  row_space: float,
  col_space: float
});

export type ContainerFormatInput = z.infer<typeof ContainerFormatInputSchema>

export const StandardInputSchema = z.object({
  dilution_factor: float,
  source: int
});

export type StandardInput = z.infer<typeof StandardInputSchema>

const bool = z.preprocess((val) => {
  if (typeof val === "string") {
    return val.toLowerCase() === "true";
  }
  return val;
}, z.boolean());

export const StandardCurveInputSchema = z.object({
  stock: z.string(),
  diluent: z.string(),
  volume: float,
  start_row: int,
  start_col: int,
  increment_along_row: bool,
  serial_dilution: bool,
  standards: z.array(StandardInputSchema)
});

export const PipBotDesignInputSchema = z.object({
  solutes: z.array(SoluteInputSchema),
  solvents: z.array(SolventInputSchema),
  stocks: z.array(StockInputSchema),
  samples: z.array(SolutionInputSchema).optional(),
  grids: z.array(ContainerFormatInputSchema),
  curve: StandardCurveInputSchema.optional(),
  dest_grid: z.string()
});

export type PipBotDesignInput = z.infer<typeof PipBotDesignInputSchema>
