# Cinema Booking API

## O problema

Múltiplos usuários precisam selecionar assentos para uma sessão de cinema. Isso precisa ocorrer de forma concorrente, sem correr o risco de dois usuários agendarem o mesmo assento, ou do cinema vender mais ingressos do que o número de assentos disponíveis.

## Possíveis soluções

### Fila única (síncrono)

Uma possível solução seria ter somente uma bilheteria e uma única fila. Assim, o primeiro da fila compra seu ingresso/assento, depois o segundo, depois o terceiro etc., e a bilheteria sempre sabe que o assento "x" já está ocupado, não podendo vender para outro comprador. O problema é que esse approach não é muito escalável. Se tivermos muitos acessos ao mesmo tempo (por exemplo, o lançamento de um filme blockbuster), levaria muito tempo para que todos os usuários pudessem fazer o booking de seus assentos.

### Múltiplas filas (paralelo)

Uma solução melhor seria ter mais de uma bilheteria (e consequentemente mais de uma fila). Todavia, essa solução adiciona um requisito novo: controlar de forma paralela os assentos que já foram escolhidos/comprados. Isso é feito com um approach "pessimista", ou seja, se uma bilheteria está processando a venda do assento "x", qualquer outra bilheteria que tentar vender esse mesmo assento precisa esperar até que a primeira venda se resolva (locking).

### Por que não usar um approach otimista?

Um approach otimista, onde não existe trava, pode até fazer sentido, porém dois usuários podem selecionar o mesmo assento, e na hora do pagamento, aquele que for confirmar por último vai ter um erro, pois o assento mudou para "ocupado" enquanto ele pagava, tornando a UX ruim.

## Abordagem escolhida

Múltiplas bilheterias (múltiplas instâncias da API) com locking pessimista por assento: nenhuma delas guarda o estado de reserva sozinha, e a concessão de um assento só é confirmada quando se garante, de forma serializada, que nenhuma outra bilheteria está processando o mesmo assento ao mesmo tempo.
