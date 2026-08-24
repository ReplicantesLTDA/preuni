"""T030: pt-BR language detection (FR-003, research R11)."""

from __future__ import annotations

import pytest

from src.corrector.prevalidation.language import LanguageMismatchError, validate_pt_br

PT_BR_SAMPLE = (
    "A inclusão digital de pessoas idosas é um desafio brasileiro contemporâneo. "
    "Você consegue perceber, no cotidiano, que muitos avós ainda têm dificuldade para utilizar "
    "aplicativos de banco e plataformas governamentais. Diante disso, é dever do Estado e da "
    "sociedade civil construir uma estratégia coordenada de inclusão, com cursos gratuitos em "
    "telecentros e parcerias com escolas técnicas para mediação digital intergeracional."
)

PT_PT_SAMPLE = (
    "A inclusão digital dos cidadãos seniores constitui um desafio contemporâneo em Portugal. "
    "Tu próprio consegues notar que muitos avós ainda têm dificuldade em utilizar os autocarros "
    "modernos, o multibanco e os serviços de e-government. Logo, é dever do Estado promover "
    "acções de literacia digital articuladas com as juntas de freguesia em todo o território."
)

EN_SAMPLE = (
    "Digital inclusion of elderly people is a pressing contemporary challenge. "
    "Many seniors still struggle to use mobile banking and government platforms in daily life. "
    "Therefore, the state and civil society must build a coordinated strategy with free courses, "
    "telecentres, and partnerships with vocational schools for intergenerational digital "
    "mediation."
)

ES_SAMPLE = (
    "La inclusión digital de las personas mayores es un desafío contemporáneo en el país. "
    "Muchos abuelos aún tienen dificultades para usar la banca móvil y las plataformas "
    "gubernamentales en la vida diaria. Por lo tanto, el Estado y la sociedad civil deben "
    "construir una estrategia coordinada con cursos gratuitos y telecentros comunitarios."
)


def test_pt_br_passes() -> None:
    validate_pt_br(PT_BR_SAMPLE)


def test_pt_pt_fails() -> None:
    with pytest.raises(LanguageMismatchError):
        validate_pt_br(PT_PT_SAMPLE)


def test_english_fails() -> None:
    with pytest.raises(LanguageMismatchError):
        validate_pt_br(EN_SAMPLE)


def test_spanish_fails() -> None:
    with pytest.raises(LanguageMismatchError):
        validate_pt_br(ES_SAMPLE)
