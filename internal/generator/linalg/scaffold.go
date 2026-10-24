package linalg

import (
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
)

// Ensure generators implement generator.ScaffoldProblem.
var (
	_ generator.ScaffoldProblem = (*eigenComputeGen)(nil)
	_ generator.ScaffoldProblem = (*diagonalizationGen)(nil)
	_ generator.ScaffoldProblem = (*cramerGen)(nil)
	_ generator.ScaffoldProblem = (*luGen)(nil)
	_ generator.ScaffoldProblem = (*qrGen)(nil)
	_ generator.ScaffoldProblem = (*svdGen)(nil)
	_ generator.ScaffoldProblem = (*jordanGen)(nil)
)

func (g *eigenComputeGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	// Representative 2x2 matrix with distinct integer eigenvalues lambda1 = 1, lambda2 = 4
	// Characteristic equation: lambda^2 - 5*lambda + 4 = 0
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "Find the trace of the matrix \\(A = \\begin{pmatrix} 2 & 2 \\\\ 1 & 3 \\end{pmatrix}\\).",
			Answer:        "5",
			Explanation:   "The trace of a matrix is the sum of its diagonal entries: 2 + 3 = 5.",
			PrereqConcept: "linalg.matrix.add",
		},
		{
			Ordinal:       2,
			Prompt:        "Find the determinant of the matrix \\(A = \\begin{pmatrix} 2 & 2 \\\\ 1 & 3 \\end{pmatrix}\\).",
			Answer:        "4",
			Explanation:   "The determinant of a 2x2 matrix is ad - bc: (2)(3) - (2)(1) = 6 - 2 = 4.",
			PrereqConcept: "linalg.det.2x2",
		},
		{
			Ordinal:       3,
			Prompt:        "Find the smaller eigenvalue from the characteristic equation \\(\\lambda^2 - 5\\lambda + 4 = 0\\).",
			Answer:        "1",
			Explanation:   "Factoring gives (lambda - 1)(lambda - 4) = 0, so the smaller eigenvalue is 1.",
			PrereqConcept: "alg.quad.formula",
		},
	}
}

func (g *diagonalizationGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For a diagonalizable matrix \\(A = PDP^{-1}\\) with eigenvalues 1, 2 and 3, compute the trace \\(\\operatorname{tr}(A)\\).",
			Answer:        "6",
			Explanation:   "The trace equals the sum of the eigenvalues: 1 + 2 + 3 = 6.",
			PrereqConcept: "linalg.eigen.compute",
		},
		{
			Ordinal:       2,
			Prompt:        "For the same matrix, compute the determinant \\(\\det(A)\\).",
			Answer:        "6",
			Explanation:   "The determinant equals the product of the eigenvalues: (1)(2)(3) = 6.",
			PrereqConcept: "linalg.det.2x2",
		},
		{
			Ordinal:       3,
			Prompt:        "What is the spectral radius \\(\\rho(A) = \\max_i |\\lambda_i|\\)?",
			Answer:        "3",
			Explanation:   "The spectral radius is the maximum eigenvalue magnitude: max(1, 2, 3) = 3.",
			PrereqConcept: "linalg.eigen.compute",
		},
	}
}

func (g *cramerGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For the linear system \\(2x + y = 5\\) and \\(x + 3y = 5\\), compute the coefficient determinant \\(\\det(A)\\).",
			Answer:        "5",
			Explanation:   "The coefficient determinant is (2)(3) - (1)(1) = 6 - 1 = 5.",
			PrereqConcept: "linalg.det.2x2",
		},
		{
			Ordinal:       2,
			Prompt:        "Compute the modified determinant \\(\\det(A_x)\\) by replacing the first column with constants \\(\\begin{pmatrix} 5 \\\\ 5 \\end{pmatrix}\\).",
			Answer:        "10",
			Explanation:   "The modified determinant is (5)(3) - (1)(5) = 15 - 5 = 10.",
			PrereqConcept: "linalg.det.2x2",
		},
		{
			Ordinal:       3,
			Prompt:        "Apply Cramer's rule \\(x = \\frac{\\det(A_x)}{\\det(A)}\\) to find \\(x\\).",
			Answer:        "2",
			Explanation:   "Dividing determinants gives x = 10 / 5 = 2.",
			PrereqConcept: "alg.systems.substitution",
		},
	}
}

func (g *luGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For \\(A = \\begin{pmatrix} 2 & 1 \\\\ 6 & 7 \\end{pmatrix}\\), what elimination multiplier \\(m_{21} = A_{21}/A_{11}\\) zeroes the entry below the pivot?",
			Answer:        "3",
			Explanation:   "Dividing entries gives m_21 = 6 / 2 = 3.",
			PrereqConcept: "linalg.matrix.mult",
		},
		{
			Ordinal:       2,
			Prompt:        "Compute the second pivot \\(U_{22} = A_{22} - m_{21} A_{12}\\).",
			Answer:        "4",
			Explanation:   "Substituting gives U_22 = 7 - 3(1) = 4.",
			PrereqConcept: "linalg.matrix.mult",
		},
		{
			Ordinal:       3,
			Prompt:        "Compute \\(\\det(A)\\) using the diagonal product of \\(U = \\begin{pmatrix} 2 & 1 \\\\ 0 & 4 \\end{pmatrix}\\).",
			Answer:        "8",
			Explanation:   "The determinant is U_11 * U_22 = 2 * 4 = 8.",
			PrereqConcept: "linalg.det.2x2",
		},
	}
}

func (g *qrGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For column vector \\(a_1 = \\begin{pmatrix} 3 \\\\ 4 \\end{pmatrix}\\), compute its Euclidean norm \\(r_{11} = \\|a_1\\|\\).",
			Answer:        "5",
			Explanation:   "The norm is sqrt(3^2 + 4^2) = sqrt(9 + 16) = sqrt(25) = 5.",
			PrereqConcept: "linalg.inner_product",
		},
		{
			Ordinal:       2,
			Prompt:        "Given unit vector \\(q_1 = \\begin{pmatrix} 3/5 \\\\ 4/5 \\end{pmatrix}\\) and \\(a_2 = \\begin{pmatrix} 5 \\\\ 0 \\end{pmatrix}\\), compute projection inner product \\(r_{12} = q_1 \\cdot a_2\\).",
			Answer:        "3",
			Explanation:   "The dot product is (3/5)(5) + (4/5)(0) = 3.",
			PrereqConcept: "linalg.inner_product",
		},
		{
			Ordinal:       3,
			Prompt:        "Subtract the projection \\(u_2 = a_2 - r_{12} q_1 = \\begin{pmatrix} 5 - 9/5 \\\\ -12/5 \\end{pmatrix} = \\begin{pmatrix} 16/5 \\\\ -12/5 \\end{pmatrix}\\). Compute norm \\(r_{22} = \\|u_2\\|\\).",
			Answer:        "4",
			Explanation:   "The norm is sqrt((16/5)^2 + (-12/5)^2) = sqrt(400/25) = sqrt(16) = 4.",
			PrereqConcept: "linalg.decomp.lu",
		},
	}
}

func (g *svdGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For diagonal matrix \\(\\Sigma = \\operatorname{diag}(4, 3)\\), compute the Gramian trace \\(\\operatorname{tr}(A^T A) = \\sigma_1^2 + \\sigma_2^2\\).",
			Answer:        "25",
			Explanation:   "Squaring and adding singular values gives 4^2 + 3^2 = 16 + 9 = 25.",
			PrereqConcept: "linalg.matrix.mult",
		},
		{
			Ordinal:       2,
			Prompt:        "Compute the Gramian determinant \\(\\det(A^T A) = \\sigma_1^2 \\sigma_2^2\\).",
			Answer:        "144",
			Explanation:   "Multiplying squared singular values gives 16 * 9 = 144.",
			PrereqConcept: "linalg.det.2x2",
		},
		{
			Ordinal:       3,
			Prompt:        "What is the condition number \\(\\kappa(A) = \\sigma_1 / \\sigma_2\\) if \\(\\sigma_1 = 6\\) and \\(\\sigma_2 = 2\\)?",
			Answer:        "3",
			Explanation:   "Dividing largest by smallest singular value gives 6 / 2 = 3.",
			PrereqConcept: "linalg.eigen.compute",
		},
	}
}

func (g *jordanGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "A 3x3 matrix has Jordan form with eigenvalue \\(\\lambda = 2\\) of algebraic multiplicity 3. What is \\(\\operatorname{tr}(A)\\)?",
			Answer:        "6",
			Explanation:   "The trace equals the sum of eigenvalues counted with algebraic multiplicity: 2 + 2 + 2 = 6.",
			PrereqConcept: "linalg.eigen.compute",
		},
		{
			Ordinal:       2,
			Prompt:        "If \\(A - 2I\\) has rank 2, what is the geometric multiplicity \\(\\dim \\ker(A - 2I)\\)?",
			Answer:        "1",
			Explanation:   "By the rank-nullity theorem, nullity = 3 - rank = 3 - 2 = 1.",
			PrereqConcept: "linalg.rank.nullity",
		},
		{
			Ordinal:       3,
			Prompt:        "With algebraic multiplicity 3 and geometric multiplicity 1, what is the size of the single Jordan block?",
			Answer:        "3",
			Explanation:   "A single eigenvector corresponds to one Jordan block encompassing the entire algebraic multiplicity of 3.",
			PrereqConcept: "linalg.decomp.svd",
		},
	}
}

// formatInt is a helper to format integer answers cleanly.
func formatInt(n int) string {
	return strconv.Itoa(n)
}
