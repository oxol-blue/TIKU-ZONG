export interface QuestionTypeOption {
  label: string;
  value: string;
}

/** 题库题型；答案始终保存为选项文字，多答案以 ### 分隔。 */
export const QUESTION_TYPES: QuestionTypeOption[] = [
  { label: "选择题", value: "single" },
  { label: "多选题", value: "multiple" },
  { label: "判断题", value: "judge" },
  { label: "简答题", value: "short_answer" },
  { label: "填空题", value: "fill_blank" },
  { label: "名词解释", value: "term_explanation" },
  { label: "论述题", value: "essay" },
  { label: "计算题", value: "calculation" },
  { label: "分录题", value: "entry" },
  { label: "资料题", value: "material" },
  { label: "连线题", value: "matching" },
  { label: "排序题", value: "ordering" },
  { label: "完型填空", value: "cloze" },
  { label: "阅读理解", value: "reading" },
  { label: "口语题", value: "speaking" },
  { label: "听力题", value: "listening" },
  { label: "共用选项题", value: "shared_options" },
  { label: "测评题", value: "assessment" },
  { label: "其它", value: "other" }
];
